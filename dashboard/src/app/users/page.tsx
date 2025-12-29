"use client"

import { useEffect, useState } from "react"
import { Header } from "@/components/layout/header"
import { UserList } from "@/components/users/user-list"
import { UserForm } from "@/components/users/user-form"
import { fetchAPI } from "@/lib/utils"

interface User {
  id: string
  email: string
  name: string
  role: string
  status: string
  groups: string[]
  created_at: string
  last_login_at?: string
}

interface UserFormData {
  email: string
  name: string
  role: string
  status: string
}

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [formOpen, setFormOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)

  async function loadUsers() {
    try {
      const data = await fetchAPI<{ users: User[] }>("/api/v1/control/users")
      setUsers(data.users || [])
    } catch (error) {
      console.error("Failed to load users:", error)
      setUsers([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadUsers()
  }, [])

  const handleCreate = () => {
    setEditingUser(null)
    setFormOpen(true)
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    setFormOpen(true)
  }

  const handleSubmit = async (data: UserFormData) => {
    try {
      if (editingUser) {
        await fetchAPI(`/api/v1/control/users/${editingUser.id}`, {
          method: "PUT",
          body: JSON.stringify({ ...data, id: editingUser.id }),
        })
      } else {
        await fetchAPI("/api/v1/control/users", {
          method: "POST",
          body: JSON.stringify(data),
        })
      }
      loadUsers()
    } catch (error) {
      console.error("Failed to save user:", error)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await fetchAPI(`/api/v1/control/users/${id}`, { method: "DELETE" })
      setUsers(users.filter(u => u.id !== id))
    } catch (error) {
      console.error("Failed to delete user:", error)
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
      <Header title="Users" description="Manage users and role-based access control" />
      <div className="flex-1 p-6">
        <UserList
          users={users}
          onCreate={handleCreate}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      </div>
      <UserForm
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleSubmit}
        initialData={editingUser || undefined}
        mode={editingUser ? "edit" : "create"}
      />
    </div>
  )
}
