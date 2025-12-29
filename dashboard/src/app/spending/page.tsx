"use client"

import { useEffect, useState } from "react"
import { Header } from "@/components/layout/header"
import { SpendingLimits } from "@/components/spending/spending-limits"
import { SpendingForm } from "@/components/spending/spending-form"
import { fetchAPI } from "@/lib/utils"

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

interface SpendingFormData {
  user_id: string
  limit_type: string
  limit_amount: number
  currency: string
  alert_at: number
}

export default function SpendingPage() {
  const [limits, setLimits] = useState<SpendingLimit[]>([])
  const [loading, setLoading] = useState(true)
  const [formOpen, setFormOpen] = useState(false)
  const [editingLimit, setEditingLimit] = useState<SpendingLimit | null>(null)

  async function loadLimits() {
    try {
      const data = await fetchAPI<{ spending_limits: SpendingLimit[] }>("/api/v1/control/spending-limits")
      setLimits(data.spending_limits || [])
    } catch (error) {
      console.error("Failed to load spending limits:", error)
      setLimits([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadLimits()
  }, [])

  const handleCreate = () => {
    setEditingLimit(null)
    setFormOpen(true)
  }

  const handleEdit = (limit: SpendingLimit) => {
    setEditingLimit(limit)
    setFormOpen(true)
  }

  const handleSubmit = async (data: SpendingFormData) => {
    try {
      if (editingLimit) {
        await fetchAPI(`/api/v1/control/spending-limits/${editingLimit.id}`, {
          method: "PUT",
          body: JSON.stringify({ ...data, id: editingLimit.id }),
        })
      } else {
        await fetchAPI("/api/v1/control/spending-limits", {
          method: "POST",
          body: JSON.stringify(data),
        })
      }
      loadLimits()
    } catch (error) {
      console.error("Failed to save spending limit:", error)
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
      <Header title="Spending Limits" description="Manage user spending limits and budgets" />
      <div className="flex-1 p-6">
        <SpendingLimits
          limits={limits}
          onCreate={handleCreate}
          onEdit={handleEdit}
        />
      </div>
      <SpendingForm
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmit}
        initialData={editingLimit || undefined}
        mode={editingLimit ? "edit" : "create"}
      />
    </div>
  )
}
