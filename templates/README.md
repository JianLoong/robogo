# Templates Directory

The `templates` directory contains financial message templates for generating standardized messages in test scenarios, particularly SWIFT (Society for Worldwide Interbank Financial Telecommunication) messages used in international banking.

## Overview

This directory provides template files used by the `swift_message` action to generate properly formatted financial messages for testing banking systems, payment processing, and financial integration scenarios.

## Directory Structure

```
templates/
└── swift/                  # SWIFT financial message templates
    └── mt103.txt          # MT103 Single Customer Credit Transfer template
```

## SWIFT Message Templates

### **MT103 - Single Customer Credit Transfer**

**Template File**: `templates/swift/mt103.txt`

**Purpose**: Generate SWIFT MT103 messages for testing single customer credit transfers (the most common SWIFT message type for customer payments).

**Template Structure**:
```
{1:F01{{.SenderBIC}}0000000000}{2:I103{{.ReceiverBIC}}N}{4:
:20:{{.TransactionRef}}
:23B:{{.BankOperationCode}}
:32A:{{.ValueDate}}{{.Currency}}{{.InterbankAmount}}
:50K:{{.OrderingCustomer}}
:59:{{.BeneficiaryCustomer}}
:71A:{{.DetailsOfCharges}}
-}
```

**Template Variables**:
- `SenderBIC`: Bank Identifier Code of the sending institution
- `ReceiverBIC`: Bank Identifier Code of the receiving institution  
- `TransactionRef`: Unique transaction reference (field :20:)
- `BankOperationCode`: Bank operation code (field :23B:, typically "CRED")
- `ValueDate`: Value date in YYMMDD format (field :32A:)
- `Currency`: Three-letter currency code (field :32A:)
- `InterbankAmount`: Transfer amount (field :32A:)
- `OrderingCustomer`: Customer initiating the transfer (field :50K:)
- `BeneficiaryCustomer`: Customer receiving the transfer (field :59:)
- `DetailsOfCharges`: Charge allocation (field :71A:, typically "SHA", "OUR", or "BEN")

## Usage with swift_message Action

### **Basic Usage**
```yaml
- name: "Generate MT103 payment message"
  action: swift_message
  args:
    - "mt103.txt"                    # Template file name
    - {                              # Data map for template variables
        SenderBIC: "BANKBEBBAXXX",
        ReceiverBIC: "BANKDEFFXXXX", 
        TransactionRef: "REF123456789",
        BankOperationCode: "CRED",
        ValueDate: "250719",
        Currency: "EUR", 
        InterbankAmount: "1000,00",
        OrderingCustomer: "John Doe\n123 Main St\nBrussels",
        BeneficiaryCustomer: "Jane Smith\n456 Elm St\nBerlin",
        DetailsOfCharges: "SHA"
      }
  result: swift_message
```

### **Dynamic Values with Current Date**
```yaml
steps:
  # Get current date in SWIFT format (YYMMDD)
  - name: "Get current date in SWIFT format"
    action: time
    args: ["060102"]  # Go time format for YYMMDD
    result: current_date
    
  # Generate unique transaction reference
  - name: "Get current timestamp for transaction ID"
    action: time
    args: ["20060102150405"]  # YYYYMMDDHHMMSS format
    result: timestamp
    
  - name: "Create unique transaction reference"
    action: string_format
    args: ["TXN{}", "${timestamp}"]
    result: transaction_ref_raw
    
  # Extract the actual formatted string
  - name: "Extract transaction reference"
    action: jq
    args: ["${transaction_ref_raw}", ".result"]
    result: transaction_ref
    
  # Generate SWIFT message with dynamic values
  - name: "Generate MT103 with dynamic date"
    action: swift_message
    args:
      - "mt103.txt"
      - {
          SenderBIC: "BANKBEBBAXXX",
          ReceiverBIC: "BANKDEFFXXXX",
          TransactionRef: "${transaction_ref}",
          BankOperationCode: "CRED",
          ValueDate: "${current_date}",        # Dynamic current date
          Currency: "EUR",
          InterbankAmount: "2500,00",
          OrderingCustomer: "Dynamic Customer\n123 Auto Street\nBrussels",
          BeneficiaryCustomer: "Test Beneficiary\n456 Generated Ave\nBerlin",
          DetailsOfCharges: "SHA"
        }
    result: dynamic_swift_message
```

### **Integration with Variables**
```yaml
variables:
  vars:
    sender_bic: "BANKBEBBAXXX"
    receiver_bic: "BANKDEFFXXXX"
    payment_amount: "5000,00"
    payment_currency: "USD"

steps:
  - name: "Generate payment with variables"
    action: swift_message
    args:
      - "mt103.txt"
      - {
          SenderBIC: "${sender_bic}",
          ReceiverBIC: "${receiver_bic}",
          TransactionRef: "TXN${ENV:TEST_RUN_ID}",
          BankOperationCode: "CRED",
          ValueDate: "250820",
          Currency: "${payment_currency}",
          InterbankAmount: "${payment_amount}",
          OrderingCustomer: "Test Corp\nTest Address",
          BeneficiaryCustomer: "Beneficiary Corp\nBeneficiary Address", 
          DetailsOfCharges: "SHA"
        }
    result: test_payment
```

## Template Engine

### **Go Template Syntax**
Templates use Go's `text/template` package with standard template syntax:

- `{{.FieldName}}`: Insert variable value
- `{{if .Condition}}...{{end}}`: Conditional inclusion
- `{{range .Array}}...{{end}}`: Loop over arrays
- `{{/* comment */}}`: Template comments

### **Template Processing**
1. **File Loading**: Template loaded from `templates/swift/filename`
2. **Variable Merging**: Data map merged with current test variables
3. **Template Execution**: Go template engine processes placeholders
4. **Message Generation**: Formatted SWIFT message returned

### **Error Handling**
The `swift_message` action provides specific error handling for:
- **Template File Not Found**: Clear message if template file missing
- **Template Parse Errors**: Syntax errors in template structure  
- **Template Execution Errors**: Missing variables or invalid data
- **Variable Type Errors**: Invalid data map format

## SWIFT Message Format

### **Message Structure**
SWIFT messages follow a standardized format with blocks:

**Block 1**: Basic Header (Sender BIC + sequence)
```
{1:F01BANKBEBBAXXX0000000000}
```

**Block 2**: Application Header (Message type + Receiver BIC)
```
{2:I103BANKDEFFXXXXN}
```

**Block 4**: Text Block (Message content with fields)
```
{4:
:20:REF123456789
:23B:CRED
:32A:250719EUR1000,00
:50K:John Doe
123 Main St
Brussels
:59:Jane Smith
456 Elm St
Berlin
:71A:SHA
-}
```

### **Field Meanings**
- **:20:** Transaction Reference Number
- **:23B:** Bank Operation Code  
- **:32A:** Value Date, Currency Code, Amount
- **:50K:** Ordering Customer (Party ordering the transfer)
- **:59:** Beneficiary Customer (Party receiving funds)
- **:71A:** Details of Charges (Who pays transfer fees)

## Testing Scenarios

### **Banking Integration Tests**
```yaml
- name: "Test payment processing workflow"
  steps:
    - name: "Generate MT103 payment"
      action: swift_message
      args: ["mt103.txt", "${payment_data}"]
      result: swift_payment
      
    - name: "Send to payment gateway"
      action: http
      args: ["POST", "${gateway_url}/payments", "${swift_payment}"]
      result: gateway_response
      
    - name: "Validate payment accepted"
      action: assert
      args: ["${gateway_response.status_code}", "==", "200"]
```

### **Message Format Validation**
```yaml
- name: "Validate SWIFT message format"
  steps:
    - name: "Generate test message"
      action: swift_message
      args: ["mt103.txt", "${test_data}"]
      result: test_message
      
    - name: "Check message starts with block 1"
      action: assert
      args: ["${test_message}", "starts_with", "{1:F01"]
      
    - name: "Check transaction reference present"
      action: assert
      args: ["${test_message}", "contains", ":20:${test_data.TransactionRef}"]
```

## Template Development

### **Creating New Templates**
1. **Research Message Type**: Understand SWIFT message specification
2. **Create Template File**: Add new `.txt` file in `templates/swift/`
3. **Define Variables**: Use `{{.VariableName}}` for dynamic fields
4. **Test Template**: Create test case with sample data
5. **Document Usage**: Add examples and field descriptions

### **Template Best Practices**
1. **Follow SWIFT Standards**: Adhere to official SWIFT message specifications
2. **Use Clear Variable Names**: Make template variables self-documenting
3. **Include Comments**: Document field purposes and formats
4. **Handle Optional Fields**: Use conditional logic for optional message fields
5. **Validate Output**: Ensure generated messages pass SWIFT validation

### **Example: Adding MT202 Template**
```yaml
# templates/swift/mt202.txt - Financial Institution Transfer
{1:F01{{.SenderBIC}}0000000000}{2:I202{{.ReceiverBIC}}N}{4:
:20:{{.TransactionRef}}
:21:{{.RelatedRef}}
:32A:{{.ValueDate}}{{.Currency}}{{.Amount}}
:53A:{{.SendersCorrespondent}}
:58A:{{.BeneficiaryInstitution}}
:72:{{.SenderToReceiverInfo}}
-}
```

## Security Considerations

### **Sensitive Data Handling**
- Use `no_log: true` when generating messages with real financial data
- Mask sensitive fields in test logs using `sensitive_fields`
- Use environment variables for production BIC codes and amounts

### **Test Data Safety**  
```yaml
- name: "Generate payment with security"
  action: swift_message
  args: ["mt103.txt", "${secure_payment_data}"]
  no_log: true                    # Suppress logging of sensitive payment data
  sensitive_fields: ["Amount", "CustomerData"]
  result: secure_payment
```

## Integration Examples

### **End-to-End Payment Test**
```yaml
testcase: "Payment Processing Integration"
description: "Test complete payment workflow with SWIFT message generation"

variables:
  vars:
    test_amount: "1000,00"
    test_currency: "EUR"
    sender_bic: "${ENV:TEST_SENDER_BIC}"
    receiver_bic: "${ENV:TEST_RECEIVER_BIC}"

steps:
  - name: "Generate MT103 payment instruction"
    action: swift_message
    args:
      - "mt103.txt"
      - {
          SenderBIC: "${sender_bic}",
          ReceiverBIC: "${receiver_bic}",
          TransactionRef: "TEST${ENV:BUILD_NUMBER}",
          BankOperationCode: "CRED",
          ValueDate: "250901", 
          Currency: "${test_currency}",
          InterbankAmount: "${test_amount}",
          OrderingCustomer: "Test Customer\nTest Address",
          BeneficiaryCustomer: "Test Beneficiary\nBeneficiary Address",
          DetailsOfCharges: "SHA"
        }
    result: payment_message
    
  - name: "Submit payment to core banking system"
    action: http
    args: ["POST", "${core_banking_url}/payments", "${payment_message}"]
    result: payment_response
    
  - name: "Extract payment ID"
    action: jq
    args: ["${payment_response}", ".body.paymentId"]
    result: payment_id
    
  - name: "Wait for payment processing"
    action: http
    args: ["GET", "${core_banking_url}/payments/${payment_id}/status"]
    extract:
      type: "jq"  
      path: ".body.status"
    result: payment_status
    retry:
      max_attempts: 10
      delay: "5s"
      backoff: "fixed"
      retry_if: "${payment_status} != 'COMPLETED'"
```

This templates directory provides essential infrastructure for testing financial systems and banking integrations using industry-standard SWIFT message formats.