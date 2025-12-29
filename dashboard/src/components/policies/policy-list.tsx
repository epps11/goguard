"use client"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { formatDate } from "@/lib/utils"
import { Edit, Trash2, Plus, Shield } from "lucide-react"

interface Policy {
  id: string
  name: string
  description: string
  type: string
  status: string
  priority: number
  created_at: string
}

interface PolicyListProps {
  policies: Policy[]
  onEdit?: (policy: Policy) => void
  onDelete?: (id: string) => void
  onCreate?: () => void
}

const statusVariant = {
  active: "success",
  inactive: "secondary",
  draft: "outline",
} as const

const typeLabels = {
  spending: "Spending Limit",
  rate_limit: "Rate Limit",
  content: "Content Filter",
  access: "Access Control",
  compliance: "Compliance",
} as const

export function PolicyList({ policies, onEdit, onDelete, onCreate }: PolicyListProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          Policies
        </CardTitle>
        {onCreate && (
          <Button onClick={onCreate} size="sm">
            <Plus className="h-4 w-4 mr-1" />
            Create Policy
          </Button>
        )}
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Priority</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {policies.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                  No policies configured. Create your first policy to get started.
                </TableCell>
              </TableRow>
            ) : (
              policies.map((policy) => (
                <TableRow key={policy.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{policy.name}</div>
                      <div className="text-sm text-muted-foreground">{policy.description}</div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {typeLabels[policy.type as keyof typeof typeLabels] || policy.type}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[policy.status as keyof typeof statusVariant] || "secondary"}>
                      {policy.status}
                    </Badge>
                  </TableCell>
                  <TableCell>{policy.priority}</TableCell>
                  <TableCell>{formatDate(policy.created_at)}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      {onEdit && (
                        <Button variant="ghost" size="icon" onClick={() => onEdit(policy)}>
                          <Edit className="h-4 w-4" />
                        </Button>
                      )}
                      {onDelete && (
                        <Button variant="ghost" size="icon" onClick={() => onDelete(policy.id)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
