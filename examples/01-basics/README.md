# Basics Examples

Fundamental operations and utilities that form the building blocks of all Robogo tests.

## Examples

### 00-util.yaml - Essential Utilities
**Complexity:** Beginner  
**Prerequisites:** None  
**Description:** Demonstrates core utility actions including UUID generation, variable manipulation, and basic logging.

**What you'll learn:**
- UUID v4 generation with the `uuid` action
- Variable setting and retrieval with the `variable` action
- Basic logging with the `log` action
- Variable substitution syntax `${variable_name}`

**Run it:**
```bash
./robogo run examples/01-basics/00-util.yaml
```

**Key features demonstrated:**
- `uuid` action for unique identifier generation
- `variable` action for dynamic variable setting
- `log` action for output and debugging
- Variable interpolation in action arguments

This example is perfect for understanding the fundamental building blocks that you'll use in more complex tests.