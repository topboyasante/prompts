ALTER TABLE downloads
  DROP CONSTRAINT downloads_prompt_id_fkey,
  ADD CONSTRAINT downloads_prompt_id_fkey
    FOREIGN KEY (prompt_id) REFERENCES prompts(id);
