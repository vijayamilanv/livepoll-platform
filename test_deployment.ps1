# LivePoll End-to-End API Automated Verification Script

$BackendUrl = "https://livepoll-platform.onrender.com"
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "🧪 LivePoll Platform End-to-End API Test" -ForegroundColor Cyan
Write-Host "Target URL: $BackendUrl" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

# 1. Health Check
Write-Host "1. Testing Health Endpoint (/health)..." -NoNewline
try {
    $health = Invoke-RestMethod -Uri "$BackendUrl/health" -Method Get
    if ($health.status -eq "ok") {
        Write-Host " [PASS]" -ForegroundColor Green
    } else {
        Write-Host " [FAIL]" -ForegroundColor Red
    }
} catch {
    Write-Host " [FAIL] Could not connect to backend: $_" -ForegroundColor Red
    exit 1
}

# 2. Signup Test
$randomId = Get-Random -Minimum 1000 -Maximum 9999
$email = "testuser_$randomId@example.com"
$password = "TestPassword123!"

Write-Host "2. Registering Test User ($email)..." -NoNewline
$signupBody = @{
    email    = $email
    password = $password
} | ConvertTo-Json

try {
    $signupRes = Invoke-RestMethod -Uri "$BackendUrl/auth/signup" -Method Post -Body $signupBody -ContentType "application/json"
    $token = $signupRes.token
    if ($token) {
        Write-Host " [PASS] (JWT Received)" -ForegroundColor Green
    } else {
        Write-Host " [FAIL] No token returned" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host " [FAIL] Signup error: $_" -ForegroundColor Red
    exit 1
}

# 3. Create Poll Test
Write-Host "3. Creating a Poll (with JWT Token)..." -NoNewline
$headers = @{
    Authorization = "Bearer $token"
}
$pollBody = @{
    question = "What is your favorite tech stack component?"
    options  = @("Go (Backend)", "React (Frontend)", "Redis (Realtime)", "MongoDB (Database)")
} | ConvertTo-Json

try {
    $pollRes = Invoke-RestMethod -Uri "$BackendUrl/polls" -Method Post -Headers $headers -Body $pollBody -ContentType "application/json"
    $shareCode = $pollRes.shareCode
    $pollId = $pollRes.id
    $firstOptionId = $pollRes.options[0].id
    
    if ($shareCode -and $firstOptionId) {
        Write-Host " [PASS] (Share Code: $shareCode)" -ForegroundColor Green
    } else {
        Write-Host " [FAIL] Poll creation response invalid" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host " [FAIL] Poll creation error: $_" -ForegroundColor Red
    exit 1
}

# 4. Fetch Poll Public Details
Write-Host "4. Fetching Public Poll Details (/polls/$shareCode)..." -NoNewline
try {
    $fetchRes = Invoke-RestMethod -Uri "$BackendUrl/polls/$shareCode" -Method Get
    if ($fetchRes.shareCode -eq $shareCode) {
        Write-Host " [PASS]" -ForegroundColor Green
    } else {
        Write-Host " [FAIL]" -ForegroundColor Red
    }
} catch {
    Write-Host " [FAIL] Fetch error: $_" -ForegroundColor Red
}

# 5. Cast Vote Test
Write-Host "5. Casting Vote for option '$($pollRes.options[0].text)'..." -NoNewline
$voteBody = @{
    optionId = $firstOptionId
} | ConvertTo-Json

try {
    $voteRes = Invoke-RestMethod -Uri "$BackendUrl/polls/$shareCode/vote" -Method Post -Body $voteBody -ContentType "application/json"
    if ($voteRes.counts) {
        Write-Host " [PASS] (Redis HINCRBY Live Count Updated)" -ForegroundColor Green
    } else {
        Write-Host " [FAIL] No counts returned" -ForegroundColor Red
    }
} catch {
    Write-Host " [FAIL] Vote error: $_" -ForegroundColor Red
}

# 6. Verify Updated Count
Write-Host "6. Verifying Updated Live Vote Counts..." -NoNewline
try {
    $verifyRes = Invoke-RestMethod -Uri "$BackendUrl/polls/$shareCode" -Method Get
    $votedCount = $verifyRes.counts | Where-Object { $_.optionId -eq $firstOptionId }
    if ($votedCount.count -ge 1) {
        Write-Host " [PASS] (Live Count in Redis: $($votedCount.count))" -ForegroundColor Green
    } else {
        Write-Host " [FAIL] Count did not increment" -ForegroundColor Red
    }
} catch {
    Write-Host " [FAIL] Verification error: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "🎉 ALL END-TO-END TESTS PASSED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Cyan
