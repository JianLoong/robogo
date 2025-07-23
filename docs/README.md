# Documentation

This directory contains technical documentation for the Robogo test automation framework.

## Available Documentation

### Architecture Documentation
- **[execution-flow-diagram.md](execution-flow-diagram.md)** - Complete Mermaid diagram showing test execution flow through the strategy pattern system

### Component Documentation
- **[../internal/README.md](../internal/README.md)** - Core architecture overview and principles
- **[../internal/actions/README.md](../internal/actions/README.md)** - Action system and all available actions
- **[../internal/execution/README.md](../internal/execution/README.md)** - Execution strategies and routing system

### User Documentation
- **[../examples/README.md](../examples/README.md)** - Comprehensive test examples and usage guide
- **[../README.md](../README.md)** - Main project README with setup and basic usage

## Documentation Structure

```
docs/
├── README.md                    # This file - documentation overview
└── execution-flow-diagram.md    # Technical architecture diagram

internal/
├── README.md                    # Core architecture principles
├── actions/README.md            # Action system documentation
└── execution/README.md          # Execution strategy documentation

examples/
└── README.md                    # User guide with test examples

README.md                        # Main project documentation
CLAUDE.md                        # Development instructions for Claude Code
```

## Documentation Philosophy

### For Developers
- **Architecture First**: Understand core principles before diving into details
- **KISS Documentation**: Clear, concise explanations without over-engineering
- **Code Examples**: Every concept illustrated with real code
- **Context Awareness**: Each README provides context for its directory

### For Users
- **Example-Driven**: Learn by seeing working test cases
- **Progressive Complexity**: From beginner to expert examples
- **Practical Focus**: Real-world scenarios and common use cases
- **Troubleshooting**: Common issues and their solutions

### For Contributors
- **Architecture Decisions**: Why things are built the way they are
- **Limitations**: Known issues and future improvements
- **Guidelines**: How to maintain consistency when adding features

## Getting Started Paths

### New Users (Want to Write Tests)
1. Start with **[../README.md](../README.md)** for installation and basic usage
2. Browse **[../examples/README.md](../examples/README.md)** for test examples
3. Reference **[../internal/actions/README.md](../internal/actions/README.md)** for available actions

### New Developers (Want to Understand Code)
1. Read **[../internal/README.md](../internal/README.md)** for architecture overview
2. Study **[execution-flow-diagram.md](execution-flow-diagram.md)** for execution flow
3. Dive into **[../internal/execution/README.md](../internal/execution/README.md)** for strategy pattern details

### Contributors (Want to Add Features)
1. Understand **[../internal/README.md](../internal/README.md)** principles (KISS, no dependency injection)
2. Review **[../internal/actions/README.md](../internal/actions/README.md)** for action patterns
3. Check **[../internal/execution/README.md](../internal/execution/README.md)** for known limitations

## Key Architecture Concepts

### KISS Principles
- **No Dependency Injection**: Direct object construction throughout
- **No Over-abstraction**: Simple, direct implementations  
- **Minimal Interfaces**: Only where absolutely necessary
- **Direct Construction**: Components create their dependencies directly

### Strategy Pattern
- **Priority-based Routing**: Higher priority strategies handle more specific cases
- **Clean Separation**: Each strategy handles one execution concern
- **Delegation**: Strategies can route back to the router for composition

### Security by Design
- **Step-level Security**: `no_log` and `sensitive_fields` properties
- **Automatic Masking**: Passwords, tokens, API keys detected and masked
- **Environment Variables**: Secure credential management via `${ENV:VAR}`

### Error Handling
- **Dual System**: ErrorInfo (technical problems) vs FailureInfo (logical test problems)
- **Structured Errors**: Category, code, template, context information
- **User-friendly**: Clear messages with actionable suggestions

## Recent Architecture Improvements

### Simplification (2024)
- **Removed 6+ abstraction layers**: VariableManager, TemplateSubstitution, ExecutionPipeline, etc.
- **Eliminated dependency injection**: Direct construction pattern throughout
- **Consolidated strategies**: Single strategy pattern handles all control flow

### File Organization (2024)
- **Split large files**: `basic_strategy.go` split into 4 focused modules
- **Single responsibility**: Each file handles one clear concern
- **Maintainable size**: No files over 300 lines

### Security Enhancements (2024)
- **Step-level security**: Security controls as step properties, not action options
- **Comprehensive masking**: JSON-aware masking with custom field support
- **No-log functionality**: Ansible-like logging suppression for sensitive operations

### SCP Support (2024)
- **Secure file transfer**: SSH/SFTP support with password and key authentication
- **Docker testing environment**: SSH servers for development testing
- **Comprehensive examples**: Upload, download, and validation test cases

## Contributing to Documentation

### Adding New Documentation
1. **Follow the pattern**: Each directory should have a README explaining its contents
2. **Link properly**: Use relative links to connect related documentation
3. **Keep current**: Update documentation when code changes
4. **Example-driven**: Include code examples for every concept

### Documentation Standards
- **Clear structure**: Use consistent headings and sections
- **Practical focus**: Prioritize what developers actually need to know
- **Accurate information**: Documentation should match actual code behavior
- **Concise writing**: Respect the reader's time

### Review Process
- **Accuracy check**: Verify all code examples work
- **Link validation**: Ensure all relative links work correctly
- **Consistency**: Follow established patterns and terminology
- **User perspective**: Consider both new and experienced users

## Maintenance

### Keeping Documentation Current
- **Code changes**: Update docs when architecture changes
- **New features**: Document new actions and capabilities
- **Bug fixes**: Update examples and troubleshooting sections
- **Regular review**: Periodic check for outdated information

### Quality Indicators
- **Examples work**: All code examples can be run successfully
- **Links valid**: No broken internal or external links
- **Up-to-date**: Reflects current codebase state
- **User feedback**: Incorporates user questions and confusion points