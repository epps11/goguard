"use client"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatDate } from "@/lib/utils"
import { AlertTriangle, Shield, DollarSign, Bell } from "lucide-react"

interface Alert {
  id: string
  type: string
  severity: string
  title: string
  message: string
  created_at: string
}

interface AlertsListProps {
  alerts: Alert[]
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

export function AlertsList({ alerts }: AlertsListProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="h-5 w-5" />
          Recent Alerts
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {alerts.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">
              No recent alerts
            </p>
          ) : (
            alerts.map((alert) => {
              const Icon = typeIcon[alert.type as keyof typeof typeIcon] || typeIcon.default
              return (
                <div key={alert.id} className="flex items-start gap-3 p-3 rounded-lg border">
                  <Icon className="h-5 w-5 mt-0.5 text-muted-foreground" />
                  <div className="flex-1 space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">{alert.title}</span>
                      <Badge variant={severityVariant[alert.severity as keyof typeof severityVariant] || "secondary"}>
                        {alert.severity}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">{alert.message}</p>
                    <p className="text-xs text-muted-foreground">{formatDate(alert.created_at)}</p>
                  </div>
                </div>
              )
            })
          )}
        </div>
      </CardContent>
    </Card>
  )
}
