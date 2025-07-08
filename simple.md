🚀 Running test case: Assert Action Test
📋 Description: Test all supported assertion operators and edge cases
📝 Steps: 19

Step 1: Assert equal numbers
✅ Basic numeric equality should work
--------------------------------------------------------------------------------
Step 2: Assert not equal numbers
✅ Different numbers should not be equal
--------------------------------------------------------------------------------
Step 3: Assert greater
✅ 2 should be greater than 1
--------------------------------------------------------------------------------
Step 4: Assert less
✅ 1 should be less than 2
--------------------------------------------------------------------------------
Step 5: Assert greater or equal
✅ 2 should be greater than or equal to 2
--------------------------------------------------------------------------------
Step 6: Assert less or equal
✅ 2 should be less than or equal to 2
--------------------------------------------------------------------------------
Step 7: Assert modulo equals
✅ ==
--------------------------------------------------------------------------------
Step 8: Assert modulo not equals
✅ !=
--------------------------------------------------------------------------------
Step 9: Assert string equal
✅ Identical strings should be equal
--------------------------------------------------------------------------------
Step 10: Assert string not equal
✅ Different strings should not be equal
--------------------------------------------------------------------------------
Step 11: Assert contains
✅ String should contain substring
--------------------------------------------------------------------------------
Step 12: Assert not contains
✅ String should not contain substring
--------------------------------------------------------------------------------
Step 13: Assert starts with
✅ String should start with prefix
--------------------------------------------------------------------------------
Step 14: Assert ends with
✅ String should end with suffix
--------------------------------------------------------------------------------
Step 15: Assert matches regex
✅ String should match alphanumeric pattern
--------------------------------------------------------------------------------
Step 16: Assert not matches regex
✅ String should not match pattern
--------------------------------------------------------------------------------
Step 17: Assert not empty string
✅ Non-empty string should not be empty
--------------------------------------------------------------------------------
Step 18: Type mismatch
❌ Assertion failed: '1' == 'one' (Numeric 1 should not equal string 'one')
❌ Step 18 failed: type=assertion | Assertion failed: '1' == 'one' (Numeric 1 should not equal string 'one')
--------------------------------------------------------------------------------
Step 19: Assert not empty string
✅ Non-empty string should not be empty
--------------------------------------------------------------------------------

## 📊 Test Results for: Assert Action Test

**❌ Status:** FAILED

**⏱️ Duration:** 591µs

**📝 Steps Summary:**

| Total  | Passed  | Failed | Skipped |
|--------|---------|--------|---------|
| 19     | 18      | 1      | 0       |


Step Results (Markdown Table):

Step Results:
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
| 19   | Assert not empty string  | assert       | ✅ PASSED   | 0s     | ✅ Non-empty string ... |                          |
