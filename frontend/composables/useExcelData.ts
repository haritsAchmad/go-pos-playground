import ExcelJS from 'exceljs'

export function useExcelData() {
  type Sheet = { name: string; columns: string[]; rows: unknown[][] }

  const downloadWorkbook = async (filename: string, sheets: Sheet[]) => {
    const workbook = new ExcelJS.Workbook()
    workbook.creator = 'Go POS Playground'
    workbook.created = new Date()
    sheets.forEach(sheet => {
      const worksheet = workbook.addWorksheet(sheet.name.slice(0, 31))
      worksheet.addRow(sheet.columns)
      sheet.rows.forEach(row => worksheet.addRow(row.map(value => value ?? '')))
      worksheet.getRow(1).font = { bold: true, color: { argb: 'FFFFFFFF' } }
      worksheet.getRow(1).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1D6B43' } }
      worksheet.views = [{ state: 'frozen', ySplit: 1 }]
      worksheet.autoFilter = { from: { row: 1, column: 1 }, to: { row: Math.max(1, worksheet.rowCount), column: Math.max(1, sheet.columns.length) } }
      worksheet.columns.forEach(column => {
        const longest = Math.max(10, ...column.values.slice(1).map(value => String(value ?? '').length))
        column.width = Math.min(longest + 2, 45)
      })
    })
    const buffer = await workbook.xlsx.writeBuffer()
    const url = URL.createObjectURL(new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }))
    const link = document.createElement('a'); link.href = url; link.download = filename; link.click(); URL.revokeObjectURL(url)
  }

  const downloadExcel = async (filename: string, columns: string[], rows: unknown[][]) => {
    await downloadWorkbook(filename, [{ name: 'Data', columns, rows }])
  }

  const parseExcel = async (file: File) => {
    const workbook = new ExcelJS.Workbook()
    await workbook.xlsx.load(new Uint8Array(await file.arrayBuffer()))
    const worksheet = workbook.worksheets[0]
    if (!worksheet || worksheet.rowCount < 2) return []

    const headers = worksheet.getRow(1).values
      .slice(1)
      .map(value => String(value ?? '').trim().toLowerCase())
    const rows: Record<string, string>[] = []
    worksheet.eachRow((row, rowNumber) => {
      if (rowNumber === 1) return
      const values = row.values.slice(1).map((_, index) => row.getCell(index + 1).text.trim())
      if (values.some(Boolean)) {
        rows.push(Object.fromEntries(headers.map((header, index) => [header, values[index] ?? ''])))
      }
    })
    return rows
  }

  return { downloadExcel, downloadWorkbook, parseExcel }
}
