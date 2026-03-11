package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/topboyasante/prompts/internal/versions"
	"github.com/topboyasante/prompts/pkg/client"
	"gopkg.in/yaml.v3"
)

type promptManifest struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Author      string `yaml:"author"`
	Inputs      []struct {
		Name     string `yaml:"name"`
		Required bool   `yaml:"required"`
	} `yaml:"inputs"`
	Tags []string `yaml:"tags"`
}

func main() {
	loadDotEnvFromParents()

	root := &cobra.Command{
		Use:   "prompt",
		Short: "Prompt registry CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureConfigDir()
		},
	}

	root.AddCommand(loginCmd(), initCmd(), publishCmd(), installCmd(), runCmd(), searchCmd())

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func loadDotEnvFromParents() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	dir := cwd
	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Overload(envPath); err != nil {
				log.Printf("cli: failed to load %s: %v", envPath, err)
			}
			return
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

func loginCmd() *cobra.Command {
	provider := "github"
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with OAuth provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := randomState(32)
			if err != nil {
				return err
			}

			provider = strings.ToLower(strings.TrimSpace(provider))
			if provider == "" {
				provider = "github"
			}

			apiBase := strings.TrimRight(os.Getenv("PROMPTS_API_URL"), "/")
			if apiBase == "" {
				apiBase = "http://localhost:8080/v1"
			}
			loginURL := fmt.Sprintf("%s/auth/%s/login?cli=true&state=%s", apiBase, provider, state)

			tokenCh := make(chan string, 1)
			errCh := make(chan error, 1)

			mux := http.NewServeMux()
			server := &http.Server{Addr: "127.0.0.1:9876", Handler: mux}
			mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
				token := r.URL.Query().Get("token")
				if token == "" {
					errCh <- fmt.Errorf("missing token in callback")
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("missing token"))
					return
				}
				if got := r.URL.Query().Get("state"); got != state {
					errCh <- fmt.Errorf("state mismatch")
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("state mismatch"))
					return
				}
				tokenCh <- token
				_, _ = w.Write([]byte("Login successful. You can close this tab."))
			})

			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- err
				}
			}()

			_ = openBrowser(loginURL)
			fmt.Println("Complete login in your browser...")

			select {
			case token := <-tokenCh:
				if err := writeToken(token); err != nil {
					return err
				}
				_ = server.Shutdown(context.Background())
				fmt.Println("Logged in successfully")
				return nil
			case err := <-errCh:
				_ = server.Shutdown(context.Background())
				return err
			case <-time.After(2 * time.Minute):
				_ = server.Shutdown(context.Background())
				return fmt.Errorf("login timed out after 2 minutes")
			}
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "github", "oauth provider: github|google")
	return cmd
}

func initCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Scaffold a prompt package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("name is required")
			}

			if err := os.MkdirAll(name, 0o755); err != nil {
				return err
			}

			files := map[string]string{
				filepath.Join(name, "prompt.yaml"): "name: " + name + "\ndescription: Describe your prompt\nversion: 1.0.0\nauthor: \ninputs:\n  - name: product\n    required: true\ntags:\n  - general\n",
				filepath.Join(name, "prompt.md"):   "Write a high-quality prompt using {{product}}.\n",
				filepath.Join(name, "README.md"):   "# " + name + "\n\nPrompt package.\n",
			}

			for path, content := range files {
				if !force {
					if _, err := os.Stat(path); err == nil {
						return fmt.Errorf("file exists: %s (use --force to overwrite)", path)
					}
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					return err
				}
			}

			fmt.Printf("Initialized prompt package in %s\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	return cmd
}

func publishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Publish current prompt package",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifest, err := loadManifest("prompt.yaml")
			if err != nil {
				return err
			}
			if err := validateManifest(manifest); err != nil {
				return err
			}

			cl, err := client.New()
			if err != nil {
				return err
			}

			prompt, status, err := cl.CreatePrompt(manifest.Name, manifest.Description, manifest.Tags)
			if err != nil {
				return err
			}
			if status == http.StatusConflict {
				return fmt.Errorf("prompt already exists; fetch prompt id and publish a new version")
			}

			rc, _, err := versions.CreateTarball(".")
			if err != nil {
				return err
			}
			defer rc.Close()
			data, err := ioReadAll(rc)
			if err != nil {
				return err
			}

			if err := cl.UploadVersion(prompt.ID, manifest.Version, data); err != nil {
				return err
			}

			fmt.Printf("Published %s@%s\n", manifest.Name, manifest.Version)
			return nil
		},
	}
}

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install [owner/name[@version]]",
		Short: "Install prompt package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := strings.TrimSpace(args[0])
			ownerName := target
			version := ""
			if strings.Contains(target, "@") {
				parts := strings.SplitN(target, "@", 2)
				ownerName, version = parts[0], parts[1]
			}
			parts := strings.Split(ownerName, "/")
			if len(parts) != 2 {
				return fmt.Errorf("expected owner/name or owner/name@version")
			}
			owner, name := parts[0], parts[1]

			cl, err := client.New()
			if err != nil {
				return err
			}

			if version == "" {
				versionsList, err := cl.GetVersions(owner, name)
				if err != nil {
					return err
				}
				if len(versionsList) == 0 {
					return fmt.Errorf("no versions found")
				}
				version = versionsList[0].Version
			}

			bytesTar, err := cl.DownloadVersion(owner, name, version)
			if err != nil {
				return err
			}

			dest := filepath.Join(".prompts", name)
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
			if err := versions.ExtractTarball(bytes.NewReader(bytesTar), dest); err != nil {
				return err
			}

			fmt.Printf("Installed %s/%s@%s into %s\n", owner, name, version, dest)
			return nil
		},
	}
}

func runCmd() *cobra.Command {
	var vars map[string]string
	cmd := &cobra.Command{
		Use:   "run [name]",
		Short: "Render prompt with variables",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			base := filepath.Join(".prompts", name)
			manifest, err := loadManifest(filepath.Join(base, "prompt.yaml"))
			if err != nil {
				return err
			}
			body, err := os.ReadFile(filepath.Join(base, "prompt.md"))
			if err != nil {
				return err
			}

			for _, in := range manifest.Inputs {
				if in.Required {
					if strings.TrimSpace(vars[in.Name]) == "" {
						return fmt.Errorf("missing required --var %s", in.Name)
					}
				}
			}

			rendered := string(body)
			for k, v := range vars {
				rendered = strings.ReplaceAll(rendered, "{{"+k+"}}", v)
			}
			fmt.Println(rendered)
			return nil
		},
	}
	cmd.Flags().StringToStringVar(&vars, "var", map[string]string{}, "template variable key=value")
	return cmd
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search prompts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := client.New()
			if err != nil {
				return err
			}
			rows, err := cl.SearchPrompts(args[0])
			if err != nil {
				return err
			}

			sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION")
			for _, row := range rows {
				_, _ = fmt.Fprintf(w, "%s\t%s\n", row.Name, row.Description)
			}
			return w.Flush()
		},
	}
}

func ensureConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(home, ".prompts"), 0o755)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func randomState(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func writeToken(token string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".prompts", "config.json")
	data, _ := json.MarshalIndent(map[string]string{"token": token}, "", "  ")
	return os.WriteFile(path, data, 0o600)
}

func loadManifest(path string) (*promptManifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m promptManifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func validateManifest(m *promptManifest) error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("manifest name is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("manifest version is required")
	}
	return nil
}

func ioReadAll(r io.Reader) ([]byte, error) { return io.ReadAll(r) }
