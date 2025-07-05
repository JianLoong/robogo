🌐 GET https://httpbin.org/get → 200 (1.0648685s)
📝 GET Response Status: {"status_code":200,"headers":{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"*","Content-Length":"273","Content-Type":"application/json","Date":"Fri, 04 Jul 2025 23:32:05 GMT","Server":"gunicorn/19.9.0"},"body":"{\n  \"args\": {}, \n  \"headers\": {\n    \"Accept-Encoding\": \"gzip\", \n    \"Host\": \"httpbin.org\", \n    \"User-Agent\": \"Go-http-client/2.0\", \n    \"X-Amzn-Trace-Id\": \"Root=1-68686475-31d1fbf9183bba976f6b25c1\"\n  }, \n  \"origin\": \"124.188.97.107\", \n  \"url\": \"https://httpbin.org/get\"\n}\n","duration":1064868500}
🌐 POST https://httpbin.org/post → 200 (252.1973ms)
📝 POST Response Status: {"status_code":200,"headers":{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"*","Content-Length":"496","Content-Type":"application/json","Date":"Fri, 04 Jul 2025 23:32:05 GMT","Server":"gunicorn/19.9.0"},"body":"{\n  \"args\": {}, \n  \"data\": \"{\\\"name\\\": \\\"Robogo\\\", \\\"version\\\": \\\"1.0\\\"}\", \n  \"files\": {}, \n  \"form\": {}, \n  \"headers\": {\n    \"Accept-Encoding\": \"gzip\", \n    \"Content-Length\": \"36\", \n    \"Content-Type\": \"application/json\", \n    \"Host\": \"httpbin.org\", \n    \"User-Agent\": \"Go-http-client/2.0\", \n    \"X-Amzn-Trace-Id\": \"Root=1-68686475-0af4e393028bdf2b2cbde99a\"\n  }, \n  \"json\": {\n    \"name\": \"Robogo\", \n    \"version\": \"1.0\"\n  }, \n  \"origin\": \"124.188.97.107\", \n  \"url\": \"https://httpbin.org/post\"\n}\n","duration":252197300}
🌐 GET https://httpbin.org/headers → 200 (255.7803ms)
📝 Headers Response: {"status_code":200,"headers":{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"*","Content-Length":"222","Content-Type":"application/json","Date":"Fri, 04 Jul 2025 23:32:05 GMT","Server":"gunicorn/19.9.0"},"body":"{\n  \"headers\": {\n    \"Accept\": \"application/json\", \n    \"Accept-Encoding\": \"gzip\", \n    \"Host\": \"httpbin.org\", \n    \"User-Agent\": \"Robogo-Test/1.0\", \n    \"X-Amzn-Trace-Id\": \"Root=1-68686475-55c3d4cb140a93ef6e5b6f0d\"\n  }\n}\n","duration":255780300}
🌐 PUT https://httpbin.org/put → 200 (254.2974ms)
📝 PUT Response: {"status_code":200,"headers":{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"*","Content-Length":"446","Content-Type":"application/json","Date":"Fri, 04 Jul 2025 23:32:06 GMT","Server":"gunicorn/19.9.0"},"body":"{\n  \"args\": {}, \n  \"data\": \"{\\\"updated\\\": true}\", \n  \"files\": {}, \n  \"form\": {}, \n  \"headers\": {\n    \"Accept-Encoding\": \"gzip\", \n    \"Content-Length\": \"17\", \n    \"Content-Type\": \"application/json\", \n    \"Host\": \"httpbin.org\", \n    \"User-Agent\": \"Go-http-client/2.0\", \n    \"X-Amzn-Trace-Id\": \"Root=1-68686476-3677b73f0c1f29b564cfb495\"\n  }, \n  \"json\": {\n    \"updated\": true\n  }, \n  \"origin\": \"124.188.97.107\", \n  \"url\": \"https://httpbin.org/put\"\n}\n","duration":254297400}
🌐 DELETE https://httpbin.org/delete → 200 (269.9844ms)
📝 DELETE Response: {"status_code":200,"headers":{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"*","Content-Length":"339","Content-Type":"application/json","Date":"Fri, 04 Jul 2025 23:32:06 GMT","Server":"gunicorn/19.9.0"},"body":"{\n  \"args\": {}, \n  \"data\": \"\", \n  \"files\": {}, \n  \"form\": {}, \n  \"headers\": {\n    \"Accept-Encoding\": \"gzip\", \n    \"Host\": \"httpbin.org\", \n    \"User-Agent\": \"Go-http-client/2.0\", \n    \"X-Amzn-Trace-Id\": \"Root=1-68686476-1da788d702f3b8e334ecadfb\"\n  }, \n  \"json\": null, \n  \"origin\": \"124.188.97.107\", \n  \"url\": \"https://httpbin.org/delete\"\n}\n","duration":269984400}
📝 Certificate Response: 
📝 PEM Certificate Response: 
📝 Custom CA Response: 
📝 PEM CA Response: 
📝 Mixed Certificate Response: 
🌐 GET https://httpbin.org/status/404 → 404 (248.5346ms)
⚠️  Response body: 
📝 Error Response: {"status_code":404,"headers":{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"*","Content-Length":"0","Content-Type":"text/html; charset=utf-8","Date":"Fri, 04 Jul 2025 23:32:07 GMT","Server":"gunicorn/19.9.0"},"body":"","duration":248534600}
# Test Results: HTTP Actions Test

## Summary
❌ **Status:** FAILED  
⏱️ **Duration:** 2.7481073s  
📝 **Steps:** 25 total, 17 passed, 8 failed

## Test Case Details
- **Name:** HTTP Actions Test
- **Description:** Test file to demonstrate HTTP request actions with various methods, options, and certificate support

## Step Results
| Step | Action | Status | Duration | Output | Error |
|------|--------|--------|----------|--------|-------|
| 1 | http_get | ✅ | 1.0648685s | {"status_code":200,"headers":{"Access... |  |
| 2 | log | ✅ | 0s | Logged: GET Response Status: {"status... |  |
| 3 | http_post | ✅ | 252.1973ms | {"status_code":200,"headers":{"Access... |  |
| 4 | log | ✅ | 0s | Logged: POST Response Status: {"statu... |  |
| 5 | http | ✅ | 255.7803ms | {"status_code":200,"headers":{"Access... |  |
| 6 | log | ✅ | 0s | Logged: Headers Response: {"status_co... |  |
| 7 | http | ✅ | 254.2974ms | {"status_code":200,"headers":{"Access... |  |
| 8 | log | ✅ | 0s | Logged: PUT Response: {"status_code":... |  |
| 9 | http | ✅ | 269.9844ms | {"status_code":200,"headers":{"Access... |  |
| 10 | log | ✅ | 0s | Logged: DELETE Response: {"status_cod... |  |
| 11 | http | ❌ | 199.827ms |  | request failed: Get "https://api.exam... |
| 12 | log | ✅ | 0s | Logged: Certificate Response:  |  |
| 13 | http | ❌ | 0s |  | request failed: Get "https://api.exam... |
| 14 | log | ✅ | 0s | Logged: PEM Certificate Response:  |  |
| 15 | http | ❌ | 202.6178ms |  | request failed: Get "https://internal... |
| 16 | log | ✅ | 0s | Logged: Custom CA Response:  |  |
| 17 | http | ❌ | 0s |  | request failed: Get "https://internal... |
| 18 | log | ✅ | 0s | Logged: PEM CA Response:  |  |
| 19 | http | ❌ | 0s |  | request failed: Post "https://secure.... |
| 20 | log | ✅ | 0s | Logged: Mixed Certificate Response:  |  |
| 21 | http_get | ✅ | 248.5346ms | {"status_code":404,"headers":{"Access... |  |
| 22 | log | ✅ | 0s | Logged: Error Response: {"status_code... |  |
| 23 | assert | ❌ | 0s |  | assertion failed: {"status_code":200,... |
| 24 | assert | ❌ | 0s |  | assertion failed: {"status_code":200,... |
| 25 | assert | ❌ | 0s |  | assertion failed: {"status_code":404,... |

## Error
❌ assertion failed: {"status_code":404,"headers":{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"*","Content-Length":"0","Content-Type":"text/html; charset=utf-8","Date":"Fri, 04 Jul 2025 23:32:07 GMT","Server":"gunicorn/19.9.0"},"body":"","duration":248534600} != 404
