"use client"

import { useEffect, useState } from "react"
import { Header } from "@/components/layout/header"
import { AuditLogs } from "@/components/audit/audit-logs"
import { fetchAPI } from "@/lib/utils"

interface AuditLog {
  id: string
  timestamp: string
  event_type: string
  action: string
  user_id: string
  user_email: string
  resource_type: string
  status: string
  ip_address: string
}

export default function AuditPage() {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function loadLogs() {
      try {
        const data = await fetchAPI<{ logs: AuditLog[] }>("/api/v1/control/audit/logs")
        setLogs(data.logs || [])
      } catch (error) {
        console.error("Failed to load audit logs:", error)
        // Demo data
        setLogs([
          { id: "1", timestamp: new Date().toISOString(), event_type: "request", action: "chat_completion", user_id: "user-001", user_email: "john@example.com", resource_type: "llm", status: "success", ip_address: "192.168.1.100" },
          { id: "2", timestamp: new Date(Date.now() - 60000).toISOString(), event_type: "security_alert", action: "injection_blocked", user_id: "user-002", user_email: "jane@example.com", resource_type: "security", status: "blocked", ip_address: "192.168.1.101" },
          { id: "3", timestamp: new Date(Date.now() - 120000).toISOString(), event_type: "policy_change", action: "policy_updated", user_id: "admin-001", user_email: "admin@example.com", resource_type: "policy", status: "success", ip_address: "10.0.0.1" },
          { id: "4", timestamp: new Date(Date.now() - 180000).toISOString(), event_type: "request", action: "chat_completion", user_id: "user-003", user_email: "bob@example.com", resource_type: "llm", status: "success", ip_address: "192.168.1.102" },
          { id: "5", timestamp: new Date(Date.now() - 240000).toISOString(), event_type: "spending_alert", action: "limit_warning", user_id: "user-001", user_email: "john@example.com", resource_type: "spending", status: "warning", ip_address: "192.168.1.100" },
        ])
      } finally {
        setLoading(false)
      }
    }
    loadLogs()
  }, [])

  const handleFilter = (query: string) => {
    console.log("Filter logs:", query)
  }

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <Header title="Audit Logs" description="View system activity and security events" />
      <div className="flex-1 p-6">
        <AuditLogs logs={logs} onFilter={handleFilter} />
      </div>
    </div>
  )
}
