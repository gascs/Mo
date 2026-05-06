import { useState } from 'react'
import { api } from '../lib/api'

export default function Tools() {
  const [exporting, setExporting] = useState(false)
  const [importing, setImporting] = useState(false)
  const [backingUp, setBackingUp] = useState(false)
  const [integrity, setIntegrity] = useState<{ status: string; result: string } | null>(null)
  const [message, setMessage] = useState('')
  const [importResult, setImportResult] = useState<any>(null)

  const handleExport = async () => {
    setExporting(true)
    setMessage('')
    try {
      await api.tools.export()
      setMessage('导出成功')
    } catch (err: any) {
      setMessage('导出失败: ' + (err.message || 'Unknown error'))
    }
    setExporting(false)
  }

  const handleImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setImporting(true)
    setMessage('')
    setImportResult(null)
    try {
      const data = await api.tools.import(file)
      setImportResult(data)
      setMessage('导入完成')
    } catch (err: any) {
      setMessage('导入失败: ' + (err.message || 'Unknown error'))
    }
    setImporting(false)
    e.target.value = ''
  }

  const handleBackup = async () => {
    setBackingUp(true)
    setMessage('')
    try {
      const res = await api.tools.backup()
      setMessage('备份完成: ' + (res as any).file)
    } catch (err: any) {
      setMessage('备份失败: ' + (err.message || 'Unknown error'))
    }
    setBackingUp(false)
  }

  const handleIntegrityCheck = async () => {
    setMessage('')
    try {
      const res = await api.tools.integrity()
      setIntegrity(res)
    } catch (err: any) {
      setMessage('检查失败: ' + (err.message || 'Unknown error'))
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold text-[var(--color-text)] mb-6">工具</h1>

      {message && (
        <div className={`mb-4 px-4 py-2 rounded-md text-sm ${
          message.includes('失败') ? 'bg-red-900/20 text-red-300 border border-red-800' : 'bg-green-900/20 text-green-300 border border-green-800'
        }`}>
          {message}
        </div>
      )}

      <div className="grid grid-cols-2 gap-6">
        {/* Export */}
        <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
          <h2 className="text-sm font-medium text-[var(--color-text)] mb-2">全站导出</h2>
          <p className="text-xs text-[var(--color-text-secondary)] mb-4">
            导出所有文章 (Markdown)、上传文件和设置为 ZIP 压缩包。
          </p>
          <button onClick={handleExport} disabled={exporting}
            className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150">
            {exporting ? '导出中...' : '导出 ZIP'}
          </button>
        </div>

        {/* Import */}
        <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
          <h2 className="text-sm font-medium text-[var(--color-text)] mb-2">导入文章</h2>
          <p className="text-xs text-[var(--color-text-secondary)] mb-4">
            上传 Markdown 文章 (含 YAML Front Matter) 的 ZIP 文件。
          </p>
          <label className={`inline-block px-4 py-2 rounded-md text-sm font-medium text-white cursor-pointer transition-all duration-150 ${
            importing ? 'opacity-50' : 'bg-[var(--color-accent)] hover:opacity-90'
          }`}>
            {importing ? '导入中...' : '上传并导入'}
            <input type="file" accept=".zip" onChange={handleImport} className="hidden" />
          </label>
          {importResult && (
            <div className="mt-3 text-xs text-[var(--color-text-secondary)] space-y-1">
              <p>文章导入: {importResult.posts_imported} 篇</p>
              <p>媒体导入: {importResult.media_imported} 个</p>
              {importResult.errors?.length > 0 && (
                <div className="mt-2 text-red-400">
                  <p>错误:</p>
                  {importResult.errors.map((e: string, i: number) => <p key={i} className="ml-2">{e}</p>)}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Backup */}
        <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
          <h2 className="text-sm font-medium text-[var(--color-text)] mb-2">手动备份</h2>
          <p className="text-xs text-[var(--color-text-secondary)] mb-4">
            立即创建数据库和文件的备份。备份保存在 backups/ 目录，最多保留 7 份。
          </p>
          <button onClick={handleBackup} disabled={backingUp}
            className="px-4 py-2 rounded-md text-sm font-medium text-white bg-[var(--color-accent)] hover:opacity-90 disabled:opacity-50 transition-all duration-150">
            {backingUp ? '备份中...' : '立即备份'}
          </button>
        </div>

        {/* Integrity Check */}
        <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-lg p-6">
          <h2 className="text-sm font-medium text-[var(--color-text)] mb-2">数据库完整性检查</h2>
          <p className="text-xs text-[var(--color-text-secondary)] mb-4">
            运行 SQLite PRAGMA integrity_check 验证数据库文件完整性。
          </p>
          <button onClick={handleIntegrityCheck}
            className="px-4 py-2 rounded-md text-sm font-medium bg-[#21262d] border border-[var(--color-border)] text-[var(--color-text)] hover:text-white hover:bg-[#30363d] transition-colors duration-150">
            检查完整性
          </button>
          {integrity && (
            <div className={`mt-3 text-xs ${integrity.status === 'healthy' ? 'text-green-400' : 'text-red-400'}`}>
              状态: {integrity.status === 'healthy' ? '正常' : '异常'} — {integrity.result}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
