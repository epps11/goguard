"use client"

import { useEffect, useState } from "react"
import { Header } from "@/components/layout/header"
import { StatsCard } from "@/components/dashboard/stats-card"
import { AlertsList } from "@/components/dashboard/alerts-list"
import { UsageChart } from "@/components/dashboard/usage-chart"
import { fetchAPI, formatCurrency, formatNumber } from "@/lib/utils"
import { Activity, Users, ShieldAlert, DollarSign } from "lucide-react"

interface DashboardMetrics {
  overview: {
    total_requests_24h: number
    requests_change_percent: number
    active_users_24h: number
    users_change_percent: number
    blocked_requests_24h: number
    blocked_change_percent: number
    total_spend_24h: number
    spend_change_percent: number
  }
  security: {
    injection_attempts_24h: number
    pii_detections_24h: number
    threats_by_level: Record<string, number>
  }
  usage: {
    requests_by_model: Record<string, number>
    requests_by_provider: Record<string, number>
  }
  recent_alerts: Array<{
    id: string
    type: string
    severity: string
    title: string
    message: string
    created_at: string
  }>
}

export default function DashboardPage() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function loadMetrics() {
      try {
        const data = await fetchAPI<DashboardMetrics>("/api/v1/control/dashboard")
        setMetrics(data)
      } catch (error) {
        console.error("Failed to load dashboard metrics:", error)
        // Use mock data for demo
        setMetrics({
          overview: {
            total_requests_24h: 12847,
            requests_change_percent: 12.5,
            active_users_24h: 234,
            users_change_percent: 5.2,
            blocked_requests_24h: 47,
            blocked_change_percent: -8.3,
            total_spend_24h: 1234.56,
            spend_change_percent: 15.7,
          },
          security: {
            injection_attempts_24h: 23,
            pii_detections_24h: 156,
            threats_by_level: { low: 12, medium: 8, high: 3 },
          },
          usage: {
            requests_by_model: { "gpt-4o": 5432, "claude-3": 3210, "gemini-pro": 2105 },
            requests_by_provider: { openai: 6500, anthropic: 4200, google: 2147 },
          },
          recent_alerts: [
            { id: "1", type: "security", severity: "high", title: "Injection Attempt Blocked", message: "Multiple injection attempts from user xyz@example.com", created_at: new Date().toISOString() },
            { id: "2", type: "spending", severity: "medium", title: "Spending Limit Warning", message: "User abc@example.com at 85% of monthly limit", created_at: new Date(Date.now() - 3600000).toISOString() },
            { id: "3", type: "policy", severity: "low", title: "Policy Updated", message: "Rate limit policy modified by admin", created_at: new Date(Date.now() - 7200000).toISOString() },
          ],
        })
      } finally {
        setLoading(false)
      }
    }
    loadMetrics()
  }, [])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <Header title="Dashboard" description="AI governance overview and metrics" />
      <div className="flex-1 p-6 space-y-6">
        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatsCard
            title="Total Requests (24h)"
            value={formatNumber(metrics?.overview.total_requests_24h || 0)}
            change={metrics?.overview.requests_change_percent}
            icon={Activity}
          />
          <StatsCard
            title="Active Users"
            value={formatNumber(metrics?.overview.active_users_24h || 0)}
            change={metrics?.overview.users_change_percent}
            icon={Users}
          />
          <StatsCard
            title="Blocked Requests"
            value={formatNumber(metrics?.overview.blocked_requests_24h || 0)}
            change={metrics?.overview.blocked_change_percent}
            icon={ShieldAlert}
          />
          <StatsCard
            title="Total Spend (24h)"
            value={formatCurrency(metrics?.overview.total_spend_24h || 0)}
            change={metrics?.overview.spend_change_percent}
            icon={DollarSign}
          />
        </div>

        {/* Charts and Alerts */}
        <div className="grid gap-6 lg:grid-cols-2">
          <UsageChart
            data={metrics?.usage.requests_by_model || {}}
            title="Requests by Model"
            type="bar"
          />
          <UsageChart
            data={metrics?.usage.requests_by_provider || {}}
            title="Requests by Provider"
            type="pie"
          />
        </div>

        {/* Alerts */}
        <AlertsList alerts={metrics?.recent_alerts || []} />
      </div>
    </div>
  )
}
