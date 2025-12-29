"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { formatCurrency, formatDate } from "@/lib/utils"
import { DollarSign, Plus, Edit } from "lucide-react"

interface SpendingLimit {
  id: string
  user_id: string
  limit_type: string
  limit_amount: number
  current_spend: number
  currency: string
  reset_at: string
  alert_at: number
}

interface SpendingLimitsProps {
  limits: SpendingLimit[]
  onEdit?: (limit: SpendingLimit) => void
  onCreate?: () => void
}

export function SpendingLimits({ limits, onEdit, onCreate }: SpendingLimitsProps) {
  const getUsagePercentage = (current: number, limit: number) => {
    return Math.min((current / limit) * 100, 100)
  }

  const getUsageColor = (percentage: number) => {
    if (percentage >= 90) return "bg-red-500"
    if (percentage >= 75) return "bg-yellow-500"
    return "bg-green-500"
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <DollarSign className="h-5 w-5" />
          Spending Limits
        </CardTitle>
        {onCreate && (
          <Button onClick={onCreate} size="sm">
            <Plus className="h-4 w-4 mr-1" />
            Add Limit
          </Button>
        )}
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>User</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Limit</TableHead>
              <TableHead>Current Spend</TableHead>
              <TableHead>Usage</TableHead>
              <TableHead>Resets</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {limits.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground py-8">
                  No spending limits configured.
                </TableCell>
              </TableRow>
            ) : (
              limits.map((limit) => {
                const percentage = getUsagePercentage(limit.current_spend, limit.limit_amount)
                return (
                  <TableRow key={limit.id}>
                    <TableCell className="font-medium">{limit.user_id}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{limit.limit_type}</Badge>
                    </TableCell>
                    <TableCell>{formatCurrency(limit.limit_amount)}</TableCell>
                    <TableCell>{formatCurrency(limit.current_spend)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className={`h-full ${getUsageColor(percentage)} transition-all`}
                            style={{ width: `${percentage}%` }}
                          />
                        </div>
                        <span className="text-sm text-muted-foreground">
                          {percentage.toFixed(0)}%
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>{formatDate(limit.reset_at)}</TableCell>
                    <TableCell className="text-right">
                      {onEdit && (
                        <Button variant="ghost" size="icon" onClick={() => onEdit(limit)}>
                          <Edit className="h-4 w-4" />
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
