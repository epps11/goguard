"use client"

import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { formatDate } from "@/lib/utils"
import { FileText, Filter } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

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

interface AuditLogsProps {
  logs: AuditLog[]
  onFilter?: (query: string) => void
}

const statusVariant = {
  success: "success",
  failure: "destructive",
  blocked: "destructive",
  warning: "warning",
} as const

const eventTypeLabels = {
  request: "API Request",
  policy_change: "Policy Change",
  user_action: "User Action",
  system_event: "System Event",
  security_alert: "Security Alert",
  spending_alert: "Spending Alert",
} as const

export function AuditLogs({ logs, onFilter }: AuditLogsProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <FileText className="h-5 w-5" />
          Audit Logs
        </CardTitle>
        <div className="flex items-center gap-2">
          <Input
            placeholder="Search logs..."
            className="w-64"
            onChange={(e) => onFilter?.(e.target.value)}
          />
          <Button variant="outline" size="icon">
            <Filter className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>Event Type</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Resource</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>IP Address</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground py-8">
                  No audit logs found.
                </TableCell>
              </TableRow>
            ) : (
              logs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell className="font-mono text-sm">
                    {formatDate(log.timestamp)}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {eventTypeLabels[log.event_type as keyof typeof eventTypeLabels] || log.event_type}
                    </Badge>
                  </TableCell>
                  <TableCell>{log.action}</TableCell>
                  <TableCell>
                    <div>
                      <div className="text-sm">{log.user_email || "—"}</div>
                      <div className="text-xs text-muted-foreground">{log.user_id || "—"}</div>
                    </div>
                  </TableCell>
                  <TableCell>{log.resource_type}</TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[log.status as keyof typeof statusVariant] || "secondary"}>
                      {log.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-sm">{log.ip_address || "—"}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
