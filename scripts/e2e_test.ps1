<#
.SYNOPSIS
  交通数据分析平台端到端测试

.DESCRIPTION
  验证两条核心链路：
  场景1：接入数据 -> 聚合任务 -> 看板更新
  场景2：触发告警条件 -> 告警记录生成 -> 告警页面数据

.PARAMETER SkipBackend
  若后端已手动启动，传 -SkipBackend 跳过自动启动

.EXAMPLE
  # 前置：docker compose up -d 启动 PostgreSQL
  docker compose up -d

  # 完整测试（自动启动后端 + 测试 + 清理）
  ./scripts/e2e_test.ps1

  # 后端已手动启动时
  ./scripts/e2e_test.ps1 -SkipBackend
#>
param(
    [switch]$SkipBackend
)

$ErrorActionPreference = "Stop"
$base = "http://localhost:8080"
$script:failed = $false
$backendProc = $null

function Pass($msg) { Write-Host "  [PASS] $msg" -ForegroundColor Green }
function Fail($msg) { Write-Host "  [FAIL] $msg" -ForegroundColor Red; $script:failed = $true }
function Assert($cond, $msg) { if ($cond) { Pass $msg } else { Fail $msg } }

# 切到项目根目录
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

# ---- 0. 前置检查 ----
Write-Host ""
Write-Host "== 前置检查 ==" -ForegroundColor Cyan
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Fail "未检测到 Go，请先安装 Go 工具链"
    exit 1
}
Pass "Go 工具链可用"

# ---- 1. 启动后端 ----
if (-not $SkipBackend) {
    Write-Host ""
    Write-Host "== 启动后端 ==" -ForegroundColor Cyan
    # 单窗口低速即触发，便于场景2确定性触发 low_speed 告警
    $env:LOW_SPEED_CONSECUTIVE = "1"
    $backendProc = Start-Process -FilePath "go" `
        -ArgumentList @("run", "./cmd/server") `
        -PassThru -WindowStyle Hidden `
        -RedirectStandardOutput "$root\e2e_backend.log" `
        -RedirectStandardError "$root\e2e_backend.err"
    Write-Host "  后端启动中（PID=$($backendProc.Id)）..."
}

# ---- 2. 等待就绪 ----
Write-Host "  等待后端就绪..."
$ready = $false
for ($i = 0; $i -lt 60; $i++) {
    try {
        $h = Invoke-RestMethod -Uri "$base/healthz" -TimeoutSec 2
        if ($h.code -eq 0) { $ready = $true; break }
    } catch {}
    Start-Sleep -Seconds 1
}
Assert $ready "后端服务已就绪"
if (-not $ready) {
    Write-Host "  后端日志（stderr）：" -ForegroundColor Yellow
    Get-Content "$root\e2e_backend.err" -ErrorAction SilentlyContinue | Select-Object -Last 20
}

if ($ready) {
    # ---- 场景1：接入数据 -> 聚合 -> 看板更新 ----
    Write-Host ""
    Write-Host "== 场景1：接入数据 -> 聚合任务 -> 看板更新 ==" -ForegroundColor Cyan

    Write-Host "  1.1 生成模拟数据..."
    $sim = Invoke-RestMethod -Method Post -Uri "$base/api/traffic/simulate" `
        -ContentType "application/json" -Body '{"intersections":5,"minutes":15,"eventsPerMinute":3}'
    Assert ($sim.code -eq 0) "模拟数据生成成功（写入 $($sim.data.successRows) 条）"

    Write-Host "  1.2 触发 1m 聚合..."
    $agg1 = Invoke-RestMethod -Method Post -Uri "$base/api/admin/aggregate/run" `
        -ContentType "application/json" -Body '{"window":"1m"}'
    Assert ($agg1.code -eq 0) "1m 聚合执行成功"

    Write-Host "  1.3 触发 5m 聚合..."
    $agg5 = Invoke-RestMethod -Method Post -Uri "$base/api/admin/aggregate/run" `
        -ContentType "application/json" -Body '{"window":"5m"}'
    Assert ($agg5.code -eq 0) "5m 聚合执行成功"

    Write-Host "  1.4 验证看板数据..."
    $overview = Invoke-RestMethod -Method Get -Uri "$base/api/dashboard/overview"
    Assert ($overview.code -eq 0) "总览接口正常"
    Assert ($overview.data.totalVehicles -gt 0) "今日总车流 > 0（实际 $($overview.data.totalVehicles)）"

    $trend = Invoke-RestMethod -Method Get -Uri "$base/api/dashboard/trend"
    Assert ($trend.code -eq 0 -and $trend.data.Count -gt 0) "趋势数据非空（$($trend.data.Count) 个点）"

    $top = Invoke-RestMethod -Method Get -Uri "$base/api/dashboard/intersections/top?limit=10"
    Assert ($top.code -eq 0 -and $top.data.Count -gt 0) "排行数据非空（$($top.data.Count) 个路口）"

    # ---- 场景2：触发告警 -> 告警生成 -> 页面显示 ----
    Write-Host ""
    Write-Host "== 场景2：触发告警条件 -> 告警记录生成 ==" -ForegroundColor Cyan

    Write-Host "  2.1 写入低速事件（speed=15 < 阈值 20，落在上一个 1m 窗口）..."
    $lowTs = (Get-Date).AddMinutes(-1).ToString("o")
    $lowBody = @{ intersectionId = "E2E-LOW-1"; timestamp = $lowTs; vehicleCount = 30; avgSpeed = 15.0; source = "e2e" } | ConvertTo-Json
    $ev = Invoke-RestMethod -Method Post -Uri "$base/api/traffic/events" -ContentType "application/json" -Body $lowBody
    Assert ($ev.code -eq 0) "低速事件写入成功"

    Write-Host "  2.2 触发 1m 聚合 + 告警评估..."
    $agg2 = Invoke-RestMethod -Method Post -Uri "$base/api/admin/aggregate/run" `
        -ContentType "application/json" -Body '{"window":"1m"}'
    Assert ($agg2.code -eq 0) "聚合 + 告警评估执行成功"

    Write-Host "  2.3 验证告警记录..."
    $alerts = Invoke-RestMethod -Method Get -Uri "$base/api/alerts"
    Assert ($alerts.code -eq 0) "告警列表接口正常"
    $lowAlerts = @($alerts.data | Where-Object { $_.ruleCode -eq "low_speed" })
    Assert ($lowAlerts.Count -gt 0) "已生成 low_speed 告警（共 $($alerts.data.Count) 条，low_speed $($lowAlerts.Count) 条）"

    if ($lowAlerts.Count -gt 0) {
        $a = $lowAlerts[0]
        Write-Host "  2.4 验证告警处理链路（id=$($a.id)）..."
        $ack = Invoke-RestMethod -Method Patch -Uri "$base/api/alerts/$($a.id)/ack"
        Assert ($ack.code -eq 0 -and $ack.data.status -eq "acked") "确认告警成功"
        $resolve = Invoke-RestMethod -Method Patch -Uri "$base/api/alerts/$($a.id)/resolve"
        Assert ($resolve.code -eq 0 -and $resolve.data.status -eq "resolved") "处理告警成功"
    }
}

# ---- 清理 ----
if ($backendProc) {
    Write-Host ""
    Write-Host "== 清理后端 ==" -ForegroundColor Cyan
    Stop-Process -Id $backendProc.Id -Force -ErrorAction SilentlyContinue
    Write-Host "  后端已停止"
}

# ---- 结果 ----
Write-Host ""
if ($script:failed) {
    Write-Host "端到端测试：存在失败项，请检查上方 [FAIL]" -ForegroundColor Red
    exit 1
} else {
    Write-Host "端到端测试：全部通过" -ForegroundColor Green
    exit 0
}
