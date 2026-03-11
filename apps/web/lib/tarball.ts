export interface TarFile {
  name: string
  content: string
}

function createHeader(name: string, size: number): Uint8Array {
  const enc = new TextEncoder()
  const header = new Uint8Array(512)

  header.set(enc.encode(name).slice(0, 100), 0)
  header.set(enc.encode('0000644\0'), 100)
  header.set(enc.encode('0000000\0'), 108)
  header.set(enc.encode('0000000\0'), 116)
  header.set(enc.encode(size.toString(8).padStart(11, '0') + '\0'), 124)
  header.set(enc.encode(Math.floor(Date.now() / 1000).toString(8).padStart(11, '0') + '\0'), 136)
  header.fill(32, 148, 156) // checksum field: spaces before computing
  header[156] = 48 // type '0' = regular file
  header.set(enc.encode('ustar\0'), 257)
  header.set(enc.encode('00'), 263)

  let sum = 0
  for (const b of header) sum += b
  header.set(enc.encode(sum.toString(8).padStart(6, '0') + '\0 '), 148)

  return header
}

function buildTar(files: TarFile[]): Uint8Array {
  const enc = new TextEncoder()
  const parts: Uint8Array[] = []

  for (const { name, content } of files) {
    const data = enc.encode(content)
    parts.push(createHeader(name, data.length))
    parts.push(data)
    const pad = (512 - (data.length % 512)) % 512
    if (pad > 0) parts.push(new Uint8Array(pad))
  }
  parts.push(new Uint8Array(1024)) // end-of-archive

  const total = parts.reduce((n, p) => n + p.length, 0)
  const out = new Uint8Array(total)
  let off = 0
  for (const p of parts) { out.set(p, off); off += p.length }
  return out
}

export async function buildTarGz(files: TarFile[]): Promise<Blob> {
  const tar = buildTar(files)
  const stream = new CompressionStream('gzip')
  const writer = stream.writable.getWriter()
  await writer.write(tar.buffer.slice(0) as ArrayBuffer)
  await writer.close()
  return new Response(stream.readable).blob()
}

export interface PromptInput {
  name: string
  required: boolean
}

export function buildPromptYaml(
  name: string,
  description: string,
  version: string,
  author: string,
  tags: string[],
  inputs: PromptInput[],
): string {
  const lines = [
    `name: ${name}`,
    `description: ${description || ''}`,
    `version: ${version}`,
  ]
  if (author) lines.push(`author: ${author}`)
  if (tags.length > 0) {
    lines.push('tags:')
    for (const tag of tags) lines.push(`  - ${tag}`)
  }
  if (inputs.length > 0) {
    lines.push('inputs:')
    for (const inp of inputs) {
      lines.push(`  - name: ${inp.name}`)
      lines.push(`    required: ${inp.required}`)
    }
  }
  return lines.join('\n') + '\n'
}
