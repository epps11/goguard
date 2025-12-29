"use client"

import { useEffect, useState } from "react"
import { Header } from "@/components/layout/header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { fetchAPI, formatDate } from "@/lib/utils"
import { Bell, Check, AlertTriangle, Shield, DollarSign } from "lucide-react"

interface Alert {
  id: string
  type: string
  severity: string
  title: string
  message: string
  user_id?: string
  policy_id?: string
  created_at: string
  acked_at?: string
  acked_by?: string
}

const severityVariant = {
  critical: "destructive",
  high: "destructive",
  medium: "warning",
  low: "secondary",
} as const

const typeIcon = {
  security: Shield,
  spending: DollarSign,
  policy: AlertTriangle,
  default: Bell,
}

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [showAcked, setShowAcked] = useState(false)

  async function loadAlerts() {
    try {
      const data = await fetchAPI<{ alerts: Alert[] }>(`/api/v1/control/alerts?include_acked=${showAcked}`)
      setAlerts(data.alerts || [])
    } catch (error) {
      console.error("Failed to load alerts:", error)
      setAlerts([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadAlerts()
  }, [showAcked])

  const handleAcknowledge = async (id: string) => {
    try {
      await fetchAPI(`/api/v1/control/alerts/${id}/ack`, { method: "POST" })
      loadAlerts()
    } catch (error) {
      console.error("Failed to acknowledge alert:", error)
    }
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
      <Header title="Alerts" description="View and manage system alerts and notifications" />
      <div className="flex-1 p-6 space-y-4">
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            <span className="font-medium">{alerts.length} Alerts</span>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowAcked(!showAcked)}
          >
            {showAcked ? "Hide Acknowledged" : "Show Acknowledged"}
          </Button>
        </div>

        <div className="space-y-4">
          {alerts.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center text-muted-foreground">
                <Bell className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No alerts to display</p>
              </CardContent>
            </Card>
          ) : (
            alerts.map((alert) => {
              const Icon = typeIcon[alert.type as keyof typeof typeIcon] || typeIcon.default
              const isAcked = !!alert.acked_at

              return (
                <Card key={alert.id} className={isAcked ? "opacity-60" : ""}>
                  <CardHeader className="pb-2">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-full ${
                          alert.severity === "critical" || alert.severity === "high"
                            ? "bg-destructive/10"
                            : alert.severity === "medium"
                            ? "bg-yellow-500/10"
                            : "bg-muted"
                        }`}>
                          <Icon className={`h-5 w-5 ${
                            alert.severity === "critical" || alert.severity === "high"
                              ? "text-destructive"
                              : alert.severity === "medium"
                              ? "text-yellow-600"
                              : "text-muted-foreground"
                          }`} />
                        </div>
                        <div>
                          <CardTitle className="text-base flex items-center gap-2">
                            {alert.title}
                            <Badge variant={severityVariant[alert.severity as keyof typeof severityVariant] || "secondary"}>
                              {alert.severity}
                            </Badge>
                          </CardTitle>
                          <p className="text-sm text-muted-foreground mt-1">
                            {formatDate(alert.created_at)}
                          </p>
                        </div>
                      </div>
                      {!isAcked && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleAcknowledge(alert.id)}
                        >
                          <Check className="h-4 w-4 mr-1" />
                          Acknowledge
                        </Button>
                      )}
                    </div>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm">{alert.message}</p>
                    {isAcked && (
                      <p className="text-xs text-muted-foreground mt-2">
                        Acknowledged {alert.acked_at && formatDate(alert.acked_at)}
                        {alert.acked_by && ` by ${alert.acked_by}`}
                      </p>
                    )}
                  </CardContent>
                </Card>
              )
            })
          )}
        </div>
      </div>
    </div>
  )
}
