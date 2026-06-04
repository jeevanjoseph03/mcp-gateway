# MCP Gateway Test Suite for PowerShell
$BASE_URL = "http://localhost:8080"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "MCP Gateway Test Suite (PowerShell)" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Health Check
Write-Host "Test 1: Gateway Health Check" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/health" -Method Get
    $response | ConvertTo-Json
    Write-Host "✓ Health check passed" -ForegroundColor Green
} catch {
    Write-Host "✗ Health check failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 2: Initialize
Write-Host "Test 2: Initialize Handshake" -ForegroundColor Yellow
$initBody = @{
    jsonrpc = "2.0"
    id = 1
    method = "initialize"
    params = @{
        protocolVersion = "2025-03-26"
        capabilities = @{}
        clientInfo = @{
            name = "test-client"
            version = "1.0.0"
        }
    }
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/mcp" -Method Post -Body $initBody -ContentType "application/json"
    $response | ConvertTo-Json -Depth 5
    if ($response.result) {
        Write-Host "✓ Initialize succeeded" -ForegroundColor Green
    } else {
        Write-Host "✗ Initialize failed" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Initialize error: $_" -ForegroundColor Red
}
Write-Host ""

# Test 3: List Tools (Federation)
Write-Host "Test 3: List Tools (Federation Test)" -ForegroundColor Yellow
$toolsBody = @{
    jsonrpc = "2.0"
    id = 2
    method = "tools/list"
    params = @{}
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/mcp" -Method Post -Body $toolsBody -ContentType "application/json"
    $toolCount = $response.result.tools.Count
    Write-Host "Found $toolCount tools" -ForegroundColor Cyan
    $response.result.tools | ForEach-Object { Write-Host "  - $($_.name)" }
    if ($toolCount -ge 2) {
        Write-Host "✓ Tools merged successfully" -ForegroundColor Green
    } else {
        Write-Host "✗ Tool merging failed (only $toolCount tools)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Tools/list error: $_" -ForegroundColor Red
}
Write-Host ""

# Test 4: Call echo tool
Write-Host "Test 4: Call 'echo' Tool" -ForegroundColor Yellow
$echoBody = @{
    jsonrpc = "2.0"
    id = 3
    method = "tools/call"
    params = @{
        name = "echo"
        arguments = @{
            message = "Hello from PowerShell test!"
        }
    }
} | ConvertTo-Json -Depth 5

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/mcp" -Method Post -Body $echoBody -ContentType "application/json"
    $result = $response.result.content[0].text
    Write-Host "Response: $result" -ForegroundColor Cyan
    Write-Host "✓ Echo tool succeeded" -ForegroundColor Green
} catch {
    Write-Host "✗ Echo tool failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 5: Call add_numbers tool
Write-Host "Test 5: Call 'add_numbers' Tool" -ForegroundColor Yellow
$addBody = @{
    jsonrpc = "2.0"
    id = 4
    method = "tools/call"
    params = @{
        name = "add_numbers"
        arguments = @{
            a = 15.5
            b = 27.3
        }
    }
} | ConvertTo-Json -Depth 5

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/mcp" -Method Post -Body $addBody -ContentType "application/json"
    $result = $response.result.content[0].text
    Write-Host "Response: $result" -ForegroundColor Cyan
    Write-Host "✓ add_numbers tool succeeded" -ForegroundColor Green
} catch {
    Write-Host "✗ add_numbers tool failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 6: Load Balancing Test
Write-Host "Test 6: Load Balancing Test" -ForegroundColor Yellow
for ($i = 1; $i -le 4; $i++) {
    $loadBody = @{
        jsonrpc = "2.0"
        id = 10 + $i
        method = "tools/call"
        params = @{
            name = "echo"
            arguments = @{
                message = "Request #$i"
            }
        }
    } | ConvertTo-Json -Depth 5
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/mcp" -Method Post -Body $loadBody -ContentType "application/json"
        $text = $response.result.content[0].text
        Write-Host "Request $i -> $text" -ForegroundColor Cyan
    } catch {
        Write-Host "Request $i failed" -ForegroundColor Red
    }
}
Write-Host ""

# Test 7: Error Handling
Write-Host "Test 7: Error Handling (Non-existent Tool)" -ForegroundColor Yellow
$errorBody = @{
    jsonrpc = "2.0"
    id = 20
    method = "tools/call"
    params = @{
        name = "non_existent_tool"
        arguments = @{}
    }
} | ConvertTo-Json -Depth 5

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/mcp" -Method Post -Body $errorBody -ContentType "application/json"
    if ($response.error) {
        Write-Host "Error returned: $($response.error.message)" -ForegroundColor Cyan
        Write-Host "✓ Proper error returned" -ForegroundColor Green
    } else {
        Write-Host "✗ Missing error for non-existent tool" -ForegroundColor Red
    }
} catch {
    $errorMsg = $_.Exception.Message
    if ($errorMsg -like "*non_existent_tool*") {
        Write-Host "✓ Error correctly handled" -ForegroundColor Green
    } else {
        Write-Host "✗ Unexpected error: $errorMsg" -ForegroundColor Red
    }
}
Write-Host ""

# Test 8: Sequential calls to test routing
Write-Host "Test 8: Sequential Tool Calls" -ForegroundColor Yellow
for ($i = 1; $i -le 3; $i++) {
    $seqBody = @{
        jsonrpc = "2.0"
        id = 100 + $i
        method = "tools/call"
        params = @{
            name = "add_numbers"
            arguments = @{
                a = $i
                b = $i * 2
            }
        }
    } | ConvertTo-Json -Depth 5
    
    try {
        $response = Invoke-RestMethod -Uri "$BASE_URL/mcp" -Method Post -Body $seqBody -ContentType "application/json"
        $result = $response.result.content[0].text
        Write-Host "Call $i: $result" -ForegroundColor Cyan
    } catch {
        Write-Host "Call $i failed" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Test Suite Complete" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan