"use client"

import { useEffect, useState } from "react"
import { Header } from "@/components/layout/header"
import { PolicyList } from "@/components/policies/policy-list"
import { PolicyForm } from "@/components/policies/policy-form"
import { fetchAPI } from "@/lib/utils"

interface Policy {
  id: string
  name: string
  description: string
  type: string
  status: string
  priority: number
  created_at: string
}

interface PolicyFormData {
  name: string
  description: string
  type: string
  status: string
  priority: number
}

export default function PoliciesPage() {
  const [policies, setPolicies] = useState<Policy[]>([])
  const [loading, setLoading] = useState(true)
  const [formOpen, setFormOpen] = useState(false)
  const [editingPolicy, setEditingPolicy] = useState<Policy | null>(null)

  async function loadPolicies() {
    try {
      const data = await fetchAPI<{ policies: Policy[] }>("/api/v1/control/policies")
      setPolicies(data.policies || [])
    } catch (error) {
      console.error("Failed to load policies:", error)
      setPolicies([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadPolicies()
  }, [])

  const handleCreate = () => {
    setEditingPolicy(null)
    setFormOpen(true)
  }

  const handleEdit = (policy: Policy) => {
    setEditingPolicy(policy)
    setFormOpen(true)
  }

  const handleSubmit = async (data: PolicyFormData) => {
    try {
      if (editingPolicy) {
        await fetchAPI(`/api/v1/control/policies/${editingPolicy.id}`, {
          method: "PUT",
          body: JSON.stringify({ ...data, id: editingPolicy.id }),
        })
      } else {
        await fetchAPI("/api/v1/control/policies", {
          method: "POST",
          body: JSON.stringify(data),
        })
      }
      loadPolicies()
    } catch (error) {
      console.error("Failed to save policy:", error)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await fetchAPI(`/api/v1/control/policies/${id}`, { method: "DELETE" })
      setPolicies(policies.filter(p => p.id !== id))
    } catch (error) {
      console.error("Failed to delete policy:", error)
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
      <Header title="Policies" description="Manage AI governance policies and rules" />
      <div className="flex-1 p-6">
        <PolicyList
          policies={policies}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      </div>
      <PolicyForm
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmit}
        initialData={editingPolicy || undefined}
        mode={editingPolicy ? "edit" : "create"}
      />
    </div>
  )
}
