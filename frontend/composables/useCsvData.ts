export function useCsvData() {
  const safeCell = (value: unknown) => {
    let text = String(value ?? '')
    if (/^[=+\-@]/.test(text)) text = `'${text}`
    return `"${text.replace(/"/g, '""')}"`
  }

  const downloadCsv = (filename: string, columns: string[], rows: unknown[][]) => {
    const csv = '\uFEFF' + [columns, ...rows].map(row => row.map(safeCell).join(',')).join('\r\n')
    const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
  }

  const parseCsv = (text: string) => {
    const rows: string[][] = []
    let row: string[] = [], cell = '', quoted = false
    const source = text.replace(/^\uFEFF/, '')
    for (let i = 0; i < source.length; i++) {
      const char = source[i]
      if (quoted) {
        if (char === '"' && source[i + 1] === '"') { cell += '"'; i++ }
        else if (char === '"') quoted = false
        else cell += char
      } else if (char === '"') quoted = true
      else if (char === ',') { row.push(cell.trim()); cell = '' }
      else if (char === '\n') { row.push(cell.trim()); if (row.some(Boolean)) rows.push(row); row = []; cell = '' }
      else if (char !== '\r') cell += char
    }
    row.push(cell.trim()); if (row.some(Boolean)) rows.push(row)
    if (rows.length < 2) return []
    const headers = rows[0].map(v => v.toLowerCase())
    return rows.slice(1).map(values => Object.fromEntries(headers.map((header, i) => [header, values[i] ?? ''])))
  }

  return { downloadCsv, parseCsv }
}
