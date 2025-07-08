🚀 Starting test suite: Complex Suite
📝 Description: A more advanced Robogo test suite with E2E test.
🔧 Running suite setup (2 steps)...
🧪 Running 4 test cases...

📋 Running test case 1/4: Assert Action Test
   Variables: 2 regular, 1 secrets
   Final merged variables for this test case:
      api_url: https://api.example.com
      default_user: suiteuser
❌ Test case failed in 521µs

📋 Running test case 2/4: Assert Action Test
   Variables: 2 regular, 1 secrets
   Final merged variables for this test case:
      api_url: https://api.example.com
      default_user: suiteuser
❌ Test case failed in 526.6µs

📋 Running test case 3/4: Assert Action Test
   Variables: 2 regular, 1 secrets
   Final merged variables for this test case:
      api_url: https://api.example.com
      default_user: suiteuser
❌ Test case failed in 1.0618ms

📋 Running test case 4/4: Kafka Timeout Error Handling
   Variables: 2 regular, 1 secrets
   Final merged variables for this test case:
      api_url: https://api.example.com
      default_user: suiteuser
✅ Test case passed in 3.0674873s
🧹 Running suite teardown (1 steps)...

============================================================
📊 Test Suite Results: Complex Suite
⏱️  Duration: 3.0728416s

## Test Case Summary
| #    | Name                     | Status   | Duration   | Error                    |
|------|--------------------------|----------|------------|--------------------------|
| 1    | Assert Action Test       | ❌ FAILED | 0.000521s  |                          |
| 2    | Assert Action Test       | ❌ FAILED | 0.0005266s |                          |
| 3    | Assert Action Test       | ❌ FAILED | 0.001062s  |                          |
| 4    | Kafka Timeout Error H... | ✅ PASSED | 3.067s     |                          |

### Step Results for Assert Action Test
| #    | Name                     | Action       | Status       | Dur    | Output                   | Error                    |
|------|--------------------------|--------------|--------------|--------|--------------------------|--------------------------|
| 1    | Assert equal numbers     | assert       | ✅ PASSED   | 0s     | ✅ Basic numeric equ... |                          |
| 2    | Assert not equal numbers | assert       | ✅ PASSED   | 0s     | ✅ Different numbers... |                          |
| 3    | Assert greater           | assert       | ✅ PASSED   | 0s     | ✅ 2 should be great... |                          |
| 4    | Assert less              | assert       | ✅ PASSED   | 0s     | ✅ 1 should be less ... |                          |
| 5    | Assert greater or equal  | assert       | ✅ PASSED   | 0s     | ✅ 2 should be great... |                          |
| 6    | Assert less or equal     | assert       | ✅ PASSED   | 0s     | ✅ 2 should be less ... |                          |
| 7    | Assert modulo equals     | assert       | ✅ PASSED   | 0s     | ✅ ==                   |                          |
| 8    | Assert modulo not equals | assert       | ✅ PASSED   | 0s     | ✅ !=                   |                          |
| 9    | Assert string equal      | assert       | ✅ PASSED   | 0s     | ✅ Identical strings... |                          |
| 10   | Assert string not equal  | assert       | ✅ PASSED   | 0s     | ✅ Different strings... |                          |
| 11   | Assert contains          | assert       | ✅ PASSED   | 0s     | ✅ String should con... |                          |
| 12   | Assert not contains      | assert       | ✅ PASSED   | 0s     | ✅ String should not... |                          |
| 13   | Assert starts with       | assert       | ✅ PASSED   | 0s     | ✅ String should sta... |                          |
| 14   | Assert ends with         | assert       | ✅ PASSED   | 0s     | ✅ String should end... |                          |
| 15   | Assert matches regex     | assert       | ✅ PASSED   | 0s     | ✅ String should mat... |                          |
| 16   | Assert not matches regex | assert       | ✅ PASSED   | 0s     | ✅ String should not... |                          |
| 17   | Assert not empty string  | assert       | ✅ PASSED   | 0s     | ✅ Non-empty string ... |                          |
| 18   | Type mismatch            | assert       | ❌ FAILED   | 0s     | <nil>                    | type=assertion | Asse... |

### Step Results for Assert Action Test
| #    | Name                     | Action       | Status       | Dur    | Output                   | Error                    |
|------|--------------------------|--------------|--------------|--------|--------------------------|--------------------------|
| 1    | Assert equal numbers     | assert       | ✅ PASSED   | 0s     | ✅ Basic numeric equ... |                          |
| 2    | Assert not equal numbers | assert       | ✅ PASSED   | 0s     | ✅ Different numbers... |                          |
| 3    | Assert greater           | assert       | ✅ PASSED   | 0s     | ✅ 2 should be great... |                          |
| 4    | Assert less              | assert       | ✅ PASSED   | 0s     | ✅ 1 should be less ... |                          |
| 5    | Assert greater or equal  | assert       | ✅ PASSED   | 0s     | ✅ 2 should be great... |                          |
| 6    | Assert less or equal     | assert       | ✅ PASSED   | 0s     | ✅ 2 should be less ... |                          |
| 7    | Assert modulo equals     | assert       | ✅ PASSED   | 0s     | ✅ ==                   |                          |
| 8    | Assert modulo not equals | assert       | ✅ PASSED   | 0s     | ✅ !=                   |                          |
| 9    | Assert string equal      | assert       | ✅ PASSED   | 0s     | ✅ Identical strings... |                          |
| 10   | Assert string not equal  | assert       | ✅ PASSED   | 0s     | ✅ Different strings... |                          |
| 11   | Assert contains          | assert       | ✅ PASSED   | 0s     | ✅ String should con... |                          |
| 12   | Assert not contains      | assert       | ✅ PASSED   | 0s     | ✅ String should not... |                          |
| 13   | Assert starts with       | assert       | ✅ PASSED   | 0s     | ✅ String should sta... |                          |
| 14   | Assert ends with         | assert       | ✅ PASSED   | 0s     | ✅ String should end... |                          |
| 15   | Assert matches regex     | assert       | ✅ PASSED   | 0s     | ✅ String should mat... |                          |
| 16   | Assert not matches regex | assert       | ✅ PASSED   | 0s     | ✅ String should not... |                          |
| 17   | Assert not empty string  | assert       | ✅ PASSED   | 0s     | ✅ Non-empty string ... |                          |
| 18   | Type mismatch            | assert       | ❌ FAILED   | 0s     | <nil>                    | type=assertion | Asse... |

### Step Results for Assert Action Test
| #    | Name                     | Action       | Status       | Dur    | Output                   | Error                    |
|------|--------------------------|--------------|--------------|--------|--------------------------|--------------------------|
| 1    | Assert equal numbers     | assert       | ✅ PASSED   | 0s     | ✅ Basic numeric equ... |                          |
| 2    | Assert not equal numbers | assert       | ✅ PASSED   | 0s     | ✅ Different numbers... |                          |
| 3    | Assert greater           | assert       | ✅ PASSED   | 0s     | ✅ 2 should be great... |                          |
| 4    | Assert less              | assert       | ✅ PASSED   | 0s     | ✅ 1 should be less ... |                          |
| 5    | Assert greater or equal  | assert       | ✅ PASSED   | 0s     | ✅ 2 should be great... |                          |
| 6    | Assert less or equal     | assert       | ✅ PASSED   | 0s     | ✅ 2 should be less ... |                          |
| 7    | Assert modulo equals     | assert       | ✅ PASSED   | 0s     | ✅ ==                   |                          |
| 8    | Assert modulo not equals | assert       | ✅ PASSED   | 0s     | ✅ !=                   |                          |
| 9    | Assert string equal      | assert       | ✅ PASSED   | 0s     | ✅ Identical strings... |                          |
| 10   | Assert string not equal  | assert       | ✅ PASSED   | 0s     | ✅ Different strings... |                          |
| 11   | Assert contains          | assert       | ✅ PASSED   | 0s     | ✅ String should con... |                          |
| 12   | Assert not contains      | assert       | ✅ PASSED   | 0s     | ✅ String should not... |                          |
| 13   | Assert starts with       | assert       | ✅ PASSED   | 0s     | ✅ String should sta... |                          |
| 14   | Assert ends with         | assert       | ✅ PASSED   | 0s     | ✅ String should end... |                          |
| 15   | Assert matches regex     | assert       | ✅ PASSED   | 0s     | ✅ String should mat... |                          |
| 16   | Assert not matches regex | assert       | ✅ PASSED   | 0s     | ✅ String should not... |                          |
| 17   | Assert not empty string  | assert       | ✅ PASSED   | 0s     | ✅ Non-empty string ... |                          |
| 18   | Type mismatch            | assert       | ❌ FAILED   | 0s     | <nil>                    | type=assertion | Asse... |

### Step Results for Kafka Timeout Error Handling
| #    | Name                     | Action       | Status       | Dur    | Output                   | Error                    |
|------|--------------------------|--------------|--------------|--------|--------------------------|--------------------------|
| 1    | Kafka Consume Timeout... | kafka        | ✅ PASSED   | 3.067s | map[error:timeout mes... | timeout                  |
| 2    | Assert Timeout Error     | assert       | ✅ PASSED   | 0s     | ✅ Expected a timeou... |                          |

## Step Summary
| Total    | Passed   | Failed   | Skipped  |
|----------|----------|----------|----------|
| 56       | 53       | 3        | 0        |

🔧 Setup: PASSED

🧹 Teardown: PASSED
============================================================
