Starting test suite: Complex Suite
Description: A more advanced Robogo test suite with E2E test.
Running suite setup (2 steps)...
[DEBUG] ExecuteTestCase: start for test case: Suite Setup
[DEBUG] RunTestCase: after ExecuteTestCase for test case: Suite Setup
[DEBUG] RunTestCase: returning normally for test case: Suite Setup
Running 2 test cases...

Running test case 2/2: Kafka Timeout Error Handling

Running test case 1/2: Get Time Action Test
   Variables: 5 regular, 1 secrets
   Variables: 2 regular, 1 secrets
   Final merged variables for this test case:
      api_url: https://api.example.com
      default_user: suiteuser
[DEBUG] ExecuteTestCase: start for test case: Get Time Action Test
   Final merged variables for this test case:
      api_url: https://api.example.com
      default_user: suiteuser
      test_message: hello from robogo
      kafka_broker: localhost:9092
      kafka_topic: robogo_test_topic
[DEBUG] ExecuteTestCase: start for test case: Kafka Timeout Error Handling
[DEBUG] RunTestCase: after ExecuteTestCase for test case: Get Time Action Test
[DEBUG] RunTestCase: returning normally for test case: Get Time Action Test
Test case passed in 0s
--------------------------------------------------------------------------------
Step 2: Assert Timeout Error
Success: Expected a timeout error when no message is available
--------------------------------------------------------------------------------
[DEBUG] executeStepsWithConfig: all steps executed for test case: Kafka Timeout Error Handling
[DEBUG] ExecuteTestCase: finished main steps for test case: Kafka Timeout Error Handling
[DEBUG] ExecuteTestCase: calculating test results for test case: Kafka Timeout Error Handling
[DEBUG] ExecuteTestCase: determining return value for test case: Kafka Timeout Error Handling
[DEBUG] ExecuteTestCase: returning for test case: Kafka Timeout Error Handling (err: <nil>)
