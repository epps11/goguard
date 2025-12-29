"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface PolicyFormData {
  name: string
  description: string
  type: string
  status: string
  priority: number
  // Type-specific configurations
  config: {
    // Spending Limit
    daily_limit?: number
    monthly_limit?: number
    currency?: string
    // Rate Limit
    requests_per_minute?: number
    requests_per_hour?: number
    burst_limit?: number
    // Content Filter
    blocked_keywords?: string
    allowed_models?: string
    max_tokens?: number
    // Access Control
    allowed_roles?: string
    allowed_users?: string
    denied_users?: string
    // Compliance
    require_audit?: boolean
    data_retention_days?: number
    pii_handling?: string
  }
}

interface PolicyFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: PolicyFormData) => void
  initialData?: Partial<PolicyFormData>
  mode?: "create" | "edit"
}

export function PolicyForm({ open, onOpenChange, onSubmit, initialData, mode = "create" }: PolicyFormProps) {
  const [formData, setFormData] = useState<PolicyFormData>({
    name: initialData?.name || "",
    description: initialData?.description || "",
    type: initialData?.type || "spending",
    status: initialData?.status || "active",
    priority: initialData?.priority || 1,
    config: initialData?.config || {},
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
    onOpenChange(false)
    setFormData({
      name: "",
      description: "",
      type: "spending",
      status: "active",
      priority: 1,
      config: {},
    })
  }

  const updateConfig = (key: string, value: string | number | boolean) => {
    setFormData({ ...formData, config: { ...formData.config, [key]: value } })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{mode === "create" ? "Create Policy" : "Edit Policy"}</DialogTitle>
            <DialogDescription>
              {mode === "create"
                ? "Create a new governance policy for AI requests."
                : "Update the policy configuration."}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Policy Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., Daily Spending Limit"
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="e.g., Limit daily spending to $50 per user"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="type">Type</Label>
                <Select
                  value={formData.type}
                  onValueChange={(value) => setFormData({ ...formData, type: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="spending">Spending Limit</SelectItem>
                    <SelectItem value="rate_limit">Rate Limit</SelectItem>
                    <SelectItem value="content">Content Filter</SelectItem>
                    <SelectItem value="access">Access Control</SelectItem>
                    <SelectItem value="compliance">Compliance</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="status">Status</Label>
                <Select
                  value={formData.status}
                  onValueChange={(value) => setFormData({ ...formData, status: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="inactive">Inactive</SelectItem>
                    <SelectItem value="draft">Draft</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="priority">Priority (1 = highest)</Label>
              <Input
                id="priority"
                type="number"
                min={1}
                max={100}
                value={formData.priority}
                onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 1 })}
              />
            </div>

            {/* Type-specific configuration */}
            {formData.type === "spending" && (
              <div className="space-y-3 p-3 bg-muted/50 rounded-lg">
                <Label className="text-sm font-medium">Spending Limit Configuration</Label>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <Label className="text-xs">Daily Limit ($)</Label>
                    <Input
                      type="number"
                      min={0}
                      value={formData.config.daily_limit || ""}
                      onChange={(e) => updateConfig("daily_limit", parseFloat(e.target.value) || 0)}
                      placeholder="e.g., 50"
                    />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-xs">Monthly Limit ($)</Label>
                    <Input
                      type="number"
                      min={0}
                      value={formData.config.monthly_limit || ""}
                      onChange={(e) => updateConfig("monthly_limit", parseFloat(e.target.value) || 0)}
                      placeholder="e.g., 500"
                    />
                  </div>
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">Currency</Label>
                  <Select
                    value={formData.config.currency || "USD"}
                    onValueChange={(v) => updateConfig("currency", v)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="USD">USD ($)</SelectItem>
                      <SelectItem value="EUR">EUR (€)</SelectItem>
                      <SelectItem value="GBP">GBP (£)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}

            {formData.type === "rate_limit" && (
              <div className="space-y-3 p-3 bg-muted/50 rounded-lg">
                <Label className="text-sm font-medium">Rate Limit Configuration</Label>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <Label className="text-xs">Requests/Minute</Label>
                    <Input
                      type="number"
                      min={1}
                      value={formData.config.requests_per_minute || ""}
                      onChange={(e) => updateConfig("requests_per_minute", parseInt(e.target.value) || 0)}
                      placeholder="e.g., 60"
                    />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-xs">Requests/Hour</Label>
                    <Input
                      type="number"
                      min={1}
                      value={formData.config.requests_per_hour || ""}
                      onChange={(e) => updateConfig("requests_per_hour", parseInt(e.target.value) || 0)}
                      placeholder="e.g., 1000"
                    />
                  </div>
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">Burst Limit</Label>
                  <Input
                    type="number"
                    min={1}
                    value={formData.config.burst_limit || ""}
                    onChange={(e) => updateConfig("burst_limit", parseInt(e.target.value) || 0)}
                    placeholder="e.g., 10"
                  />
                </div>
              </div>
            )}

            {formData.type === "content" && (
              <div className="space-y-3 p-3 bg-muted/50 rounded-lg">
                <Label className="text-sm font-medium">Content Filter Configuration</Label>
                <div className="space-y-1">
                  <Label className="text-xs">Blocked Keywords (comma-separated)</Label>
                  <Input
                    value={formData.config.blocked_keywords || ""}
                    onChange={(e) => updateConfig("blocked_keywords", e.target.value)}
                    placeholder="e.g., password, secret, confidential"
                  />
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">Allowed Models (comma-separated)</Label>
                  <Input
                    value={formData.config.allowed_models || ""}
                    onChange={(e) => updateConfig("allowed_models", e.target.value)}
                    placeholder="e.g., gpt-4o, claude-3-5-sonnet"
                  />
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">Max Tokens per Request</Label>
                  <Input
                    type="number"
                    min={1}
                    value={formData.config.max_tokens || ""}
                    onChange={(e) => updateConfig("max_tokens", parseInt(e.target.value) || 0)}
                    placeholder="e.g., 4096"
                  />
                </div>
              </div>
            )}

            {formData.type === "access" && (
              <div className="space-y-3 p-3 bg-muted/50 rounded-lg">
                <Label className="text-sm font-medium">Access Control Configuration</Label>
                <div className="space-y-1">
                  <Label className="text-xs">Allowed Roles (comma-separated)</Label>
                  <Input
                    value={formData.config.allowed_roles || ""}
                    onChange={(e) => updateConfig("allowed_roles", e.target.value)}
                    placeholder="e.g., admin, user"
                  />
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">Allowed Users (comma-separated emails)</Label>
                  <Input
                    value={formData.config.allowed_users || ""}
                    onChange={(e) => updateConfig("allowed_users", e.target.value)}
                    placeholder="e.g., john@example.com, jane@example.com"
                  />
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">Denied Users (comma-separated emails)</Label>
                  <Input
                    value={formData.config.denied_users || ""}
                    onChange={(e) => updateConfig("denied_users", e.target.value)}
                    placeholder="e.g., blocked@example.com"
                  />
                </div>
              </div>
            )}

            {formData.type === "compliance" && (
              <div className="space-y-3 p-3 bg-muted/50 rounded-lg">
                <Label className="text-sm font-medium">Compliance Configuration</Label>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <Label className="text-xs">Require Audit Logging</Label>
                    <Select
                      value={formData.config.require_audit ? "true" : "false"}
                      onValueChange={(v) => updateConfig("require_audit", v === "true")}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="true">Yes</SelectItem>
                        <SelectItem value="false">No</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1">
                    <Label className="text-xs">Data Retention (days)</Label>
                    <Input
                      type="number"
                      min={1}
                      value={formData.config.data_retention_days || ""}
                      onChange={(e) => updateConfig("data_retention_days", parseInt(e.target.value) || 0)}
                      placeholder="e.g., 90"
                    />
                  </div>
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">PII Handling</Label>
                  <Select
                    value={formData.config.pii_handling || "mask"}
                    onValueChange={(v) => updateConfig("pii_handling", v)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="mask">Mask PII</SelectItem>
                      <SelectItem value="block">Block if PII detected</SelectItem>
                      <SelectItem value="allow">Allow (log only)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit">
              {mode === "create" ? "Create Policy" : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
