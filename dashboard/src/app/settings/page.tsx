"use client"

import { useState, useEffect } from "react"
import { Header } from "@/components/layout/header"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Settings, Shield, Bell, Database, Key, Loader2 } from "lucide-react"
import { fetchAPI } from "@/lib/utils"

export default function SettingsPage() {
  const [llmProvider, setLlmProvider] = useState("openai")
  const [llmModel, setLlmModel] = useState("gpt-4o")
  const [customModel, setCustomModel] = useState("")
  const [useCustomModel, setUseCustomModel] = useState(false)
  const [apiKey, setApiKey] = useState("")
  const [baseUrl, setBaseUrl] = useState("")
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(true)

  // Security settings
  const [injectionDetection, setInjectionDetection] = useState(true)
  const [blockOnDetection, setBlockOnDetection] = useState(true)
  const [piiMasking, setPiiMasking] = useState(true)
  const [rateLimit, setRateLimit] = useState(100)

  useEffect(() => {
    async function loadSettings() {
      try {
        const llmSettings = await fetchAPI<{
          provider: string
          model: string
          base_url?: string
        }>("/api/v1/control/settings/llm")
        if (llmSettings.provider) setLlmProvider(llmSettings.provider)
        if (llmSettings.model) setLlmModel(llmSettings.model)
        if (llmSettings.base_url) setBaseUrl(llmSettings.base_url)

        const secSettings = await fetchAPI<{
          injection_detection_enabled: boolean
          block_on_detection: boolean
          pii_masking_enabled: boolean
          rate_limit_per_minute: number
        }>("/api/v1/control/settings/security")
        if (secSettings.injection_detection_enabled !== undefined)
          setInjectionDetection(secSettings.injection_detection_enabled)
        if (secSettings.block_on_detection !== undefined)
          setBlockOnDetection(secSettings.block_on_detection)
        if (secSettings.pii_masking_enabled !== undefined)
          setPiiMasking(secSettings.pii_masking_enabled)
        if (secSettings.rate_limit_per_minute)
          setRateLimit(secSettings.rate_limit_per_minute)
      } catch (error) {
        console.error("Failed to load settings:", error)
      } finally {
        setLoading(false)
      }
    }
    loadSettings()
  }, [])

  const handleSave = async () => {
    setSaving(true)
    try {
      // Save LLM settings
      await fetchAPI("/api/v1/control/settings/llm", {
        method: "PUT",
        body: JSON.stringify({
          provider: llmProvider,
          model: useCustomModel ? customModel : llmModel,
          api_key: apiKey || undefined,
          base_url: baseUrl || undefined,
        }),
      })

      // Save security settings
      await fetchAPI("/api/v1/control/settings/security", {
        method: "PUT",
        body: JSON.stringify({
          injection_detection_enabled: injectionDetection,
          block_on_detection: blockOnDetection,
          pii_masking_enabled: piiMasking,
          rate_limit_per_minute: rateLimit,
        }),
      })

      alert("Settings saved successfully!")
    } catch (error) {
      console.error("Failed to save settings:", error)
      alert("Failed to save settings. Check console for details.")
    } finally {
      setSaving(false)
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
      <Header title="Settings" description="Configure GoGuard system settings" />
      <div className="flex-1 p-6 space-y-6">
        {/* LLM Configuration */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Key className="h-5 w-5" />
              LLM Configuration
            </CardTitle>
            <CardDescription>
              Configure the AI provider and model for processing requests
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="provider">Provider</Label>
                <Select value={llmProvider} onValueChange={setLlmProvider}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select provider" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="openai">OpenAI</SelectItem>
                    <SelectItem value="anthropic">Anthropic (Claude)</SelectItem>
                    <SelectItem value="google">Google (Gemini)</SelectItem>
                    <SelectItem value="bedrock">AWS Bedrock</SelectItem>
                    <SelectItem value="ollama">Ollama (Local)</SelectItem>
                    <SelectItem value="xai">X.AI (Grok)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="model">Model</Label>
                {!useCustomModel ? (
                  <Select value={llmModel} onValueChange={setLlmModel}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select model" />
                    </SelectTrigger>
                    <SelectContent>
                      {llmProvider === "openai" && (
                        <>
                          <SelectItem value="gpt-4o">GPT-4o</SelectItem>
                          <SelectItem value="gpt-4o-mini">GPT-4o Mini</SelectItem>
                          <SelectItem value="o1">o1</SelectItem>
                          <SelectItem value="o1-mini">o1 Mini</SelectItem>
                          <SelectItem value="o1-pro">o1 Pro</SelectItem>
                        </>
                      )}
                      {llmProvider === "anthropic" && (
                        <>
                          <SelectItem value="claude-sonnet-4">Claude Sonnet 4</SelectItem>
                          <SelectItem value="claude-3-7-sonnet">Claude 3.7 Sonnet</SelectItem>
                          <SelectItem value="claude-3-5-haiku">Claude 3.5 Haiku</SelectItem>
                          <SelectItem value="claude-3-5-sonnet">Claude 3.5 Sonnet</SelectItem>
                        </>
                      )}
                      {llmProvider === "google" && (
                        <>
                          <SelectItem value="gemini-2.5-pro">Gemini 2.5 Pro</SelectItem>
                          <SelectItem value="gemini-2.5-flash">Gemini 2.5 Flash</SelectItem>
                          <SelectItem value="gemini-2.0-flash">Gemini 2.0 Flash</SelectItem>
                        </>
                      )}
                      {llmProvider === "ollama" && (
                        <>
                          <SelectItem value="llama3.3">Llama 3.3</SelectItem>
                          <SelectItem value="llama3.2">Llama 3.2</SelectItem>
                          <SelectItem value="mistral-large">Mistral Large</SelectItem>
                          <SelectItem value="qwen2.5">Qwen 2.5</SelectItem>
                          <SelectItem value="deepseek-r1">DeepSeek R1</SelectItem>
                        </>
                      )}
                      {llmProvider === "bedrock" && (
                        <>
                          <SelectItem value="anthropic.claude-3-5-sonnet-20241022-v2:0">Claude 3.5 Sonnet v2</SelectItem>
                          <SelectItem value="anthropic.claude-3-5-haiku-20241022-v1:0">Claude 3.5 Haiku</SelectItem>
                          <SelectItem value="amazon.nova-pro-v1:0">Amazon Nova Pro</SelectItem>
                          <SelectItem value="amazon.nova-lite-v1:0">Amazon Nova Lite</SelectItem>
                          <SelectItem value="meta.llama3-2-90b-instruct-v1:0">Llama 3.2 90B</SelectItem>
                          <SelectItem value="mistral.mistral-large-2411-v1:0">Mistral Large</SelectItem>
                        </>
                      )}
                      {llmProvider === "xai" && (
                        <>
                          <SelectItem value="grok-3">Grok 3</SelectItem>
                          <SelectItem value="grok-3-mini">Grok 3 Mini</SelectItem>
                        </>
                      )}
                    </SelectContent>
                  </Select>
                ) : (
                  <Input
                    value={customModel}
                    onChange={(e) => setCustomModel(e.target.value)}
                    placeholder="e.g., gpt-4-turbo-2024-04-09"
                  />
                )}
                <Button
                  type="button"
                  variant="link"
                  size="sm"
                  className="h-auto p-0 text-xs"
                  onClick={() => setUseCustomModel(!useCustomModel)}
                >
                  {useCustomModel ? "← Use preset models" : "Use custom model name"}
                </Button>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="apiKey">API Key</Label>
                <Input
                  id="apiKey"
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="Enter your API key"
                />
                <p className="text-xs text-muted-foreground">
                  Stored securely, never exposed in logs
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="baseUrl">Base URL (Optional)</Label>
                <Input
                  id="baseUrl"
                  value={baseUrl}
                  onChange={(e) => setBaseUrl(e.target.value)}
                  placeholder="e.g., https://api.openai.com/v1"
                />
                <p className="text-xs text-muted-foreground">
                  Custom endpoint for self-hosted or proxy
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Security Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              Security Settings
            </CardTitle>
            <CardDescription>
              Configure injection detection and content filtering
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Injection Detection</Label>
                <Select
                  value={injectionDetection ? "enabled" : "disabled"}
                  onValueChange={(v) => setInjectionDetection(v === "enabled")}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="enabled">Enabled</SelectItem>
                    <SelectItem value="disabled">Disabled</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Block on Detection</Label>
                <Select
                  value={blockOnDetection ? "true" : "false"}
                  onValueChange={(v) => setBlockOnDetection(v === "true")}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="true">Block Request</SelectItem>
                    <SelectItem value="false">Allow with Warning</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>PII Masking</Label>
                <Select
                  value={piiMasking ? "enabled" : "disabled"}
                  onValueChange={(v) => setPiiMasking(v === "enabled")}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="enabled">Enabled</SelectItem>
                    <SelectItem value="disabled">Disabled</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Rate Limit (req/min)</Label>
                <Input
                  type="number"
                  value={rateLimit}
                  onChange={(e) => setRateLimit(parseInt(e.target.value) || 100)}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              Notification Settings
            </CardTitle>
            <CardDescription>
              Configure alert notifications and webhooks
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>Webhook URL</Label>
              <Input placeholder="https://your-webhook-url.com/alerts" />
            </div>
            <div className="space-y-2">
              <Label>Email Notifications</Label>
              <Input placeholder="admin@example.com, security@example.com" />
              <p className="text-xs text-muted-foreground">
                Comma-separated list of email addresses for alert notifications
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Database Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              Storage Settings
            </CardTitle>
            <CardDescription>
              Current storage configuration (read-only)
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-muted-foreground">Storage Type:</span>
                <span className="ml-2 font-medium">In-Memory</span>
              </div>
              <div>
                <span className="text-muted-foreground">Audit Log Retention:</span>
                <span className="ml-2 font-medium">10,000 entries</span>
              </div>
            </div>
            <p className="text-xs text-muted-foreground mt-4">
              Note: Data is stored in-memory and will be lost on restart. Configure a database for persistent storage.
            </p>
          </CardContent>
        </Card>

        <div className="flex justify-end">
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Settings className="h-4 w-4 mr-2" />
            )}
            {saving ? "Saving..." : "Save Settings"}
          </Button>
        </div>
      </div>
    </div>
  )
}
