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

interface SpendingFormData {
  user_id: string
  limit_type: string
  limit_amount: number
  currency: string
  alert_at: number
}

interface SpendingFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: SpendingFormData) => void
  initialData?: Partial<SpendingFormData>
  mode?: "create" | "edit"
}

export function SpendingForm({ open, onOpenChange, onSubmit, initialData, mode = "create" }: SpendingFormProps) {
  const [formData, setFormData] = useState<SpendingFormData>({
    user_id: initialData?.user_id || "",
    limit_type: initialData?.limit_type || "daily",
    limit_amount: initialData?.limit_amount || 50,
    currency: initialData?.currency || "USD",
    alert_at: initialData?.alert_at || 80,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
    onOpenChange(false)
    setFormData({
      user_id: "",
      limit_type: "daily",
      limit_amount: 50,
      currency: "USD",
      alert_at: 80,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{mode === "create" ? "Add Spending Limit" : "Edit Spending Limit"}</DialogTitle>
            <DialogDescription>
              {mode === "create"
                ? "Set a spending limit for a user or group."
                : "Update the spending limit configuration."}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="user_id">User ID or Email</Label>
              <Input
                id="user_id"
                value={formData.user_id}
                onChange={(e) => setFormData({ ...formData, user_id: e.target.value })}
                placeholder="e.g., john@example.com"
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="limit_type">Limit Type</Label>
                <Select
                  value={formData.limit_type}
                  onValueChange={(value) => setFormData({ ...formData, limit_type: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="daily">Daily</SelectItem>
                    <SelectItem value="weekly">Weekly</SelectItem>
                    <SelectItem value="monthly">Monthly</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="currency">Currency</Label>
                <Select
                  value={formData.currency}
                  onValueChange={(value) => setFormData({ ...formData, currency: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select currency" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="USD">USD ($)</SelectItem>
                    <SelectItem value="EUR">EUR (€)</SelectItem>
                    <SelectItem value="GBP">GBP (£)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="limit_amount">Limit Amount</Label>
                <Input
                  id="limit_amount"
                  type="number"
                  min={1}
                  step={0.01}
                  value={formData.limit_amount}
                  onChange={(e) => setFormData({ ...formData, limit_amount: parseFloat(e.target.value) || 0 })}
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="alert_at">Alert at (%)</Label>
                <Input
                  id="alert_at"
                  type="number"
                  min={1}
                  max={100}
                  value={formData.alert_at}
                  onChange={(e) => setFormData({ ...formData, alert_at: parseInt(e.target.value) || 80 })}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit">
              {mode === "create" ? "Add Limit" : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
