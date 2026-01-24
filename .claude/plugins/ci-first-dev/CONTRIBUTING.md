# Contributing to This Plugin

Thank you for your interest in contributing! This document provides guidelines for contributing to the plugin.

## Getting Started

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/yourusername/plugin-name.git
   cd plugin-name
   ```
3. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Claude Code CLI installed
- Git for version control
- Node.js (if using JavaScript-based MCP servers)
- Python 3.x (if using Python-based tools)

### Local Testing

1. Create a test marketplace (see README.md)
2. Install the plugin locally
3. Test all components thoroughly
4. Verify hooks and MCP servers work correctly

## Making Changes

### Code Style

- Use consistent indentation (2 spaces for JSON, 4 for Python)
- Write clear, descriptive comments
- Follow existing patterns in the codebase
- Keep functions focused and single-purpose

### Component Guidelines

#### Slash Commands
- Use descriptive names (kebab-case)
- Include clear descriptions
- Provide argument hints
- Document expected behavior
- Add usage examples

#### Agents
- Single responsibility principle
- Clear descriptions of when to use
- Specific tool restrictions
- Comprehensive instructions
- Example workflows

#### Skills
- Focused capabilities
- Clear invocation criteria
- Progressive disclosure (use REFERENCE.md and EXAMPLES.md)
- Helper scripts when needed

#### Hooks
- Non-blocking by default (exit 0)
- Clear error messages
- Graceful degradation
- Minimal performance impact

### Documentation

- Update README.md if adding new features
- Add entries to CHANGELOG.md
- Document new commands, agents, or skills
- Include usage examples
- Update this CONTRIBUTING.md if changing processes

### Commit Messages

Follow the conventional commits format:

```
type(scope): brief description

Detailed explanation if needed

Breaking changes or special notes
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Formatting, missing semicolons, etc.
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(commands): add deployment automation command

fix(hooks): format-code script now handles Unicode

docs(readme): update installation instructions
```

## Testing Checklist

Before submitting:

- [ ] All commands work as expected
- [ ] Agents can be invoked correctly
- [ ] Skills are discoverable and functional
- [ ] Hooks execute without errors
- [ ] Scripts are executable (`chmod +x`)
- [ ] No secrets or credentials in code
- [ ] Documentation is updated
- [ ] CHANGELOG.md has new entry
- [ ] Code follows existing patterns
- [ ] No unnecessary dependencies added

## Pull Request Process

1. **Ensure tests pass**
   - Test all components locally
   - Verify no regressions

2. **Update documentation**
   - README.md if needed
   - CHANGELOG.md with changes
   - Inline comments for complex code

3. **Create pull request**
   - Clear title describing the change
   - Detailed description of what and why
   - Reference any related issues
   - Include screenshots/examples if relevant

4. **PR Description Template**
   ```markdown
   ## Description
   Brief description of changes

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Changes Made
   - Change 1
   - Change 2
   - Change 3

   ## Testing
   How you tested these changes

   ## Checklist
   - [ ] Code follows style guidelines
   - [ ] Documentation updated
   - [ ] CHANGELOG.md updated
   - [ ] All tests pass
   - [ ] No new warnings
   ```

5. **Review process**
   - Maintainers will review your PR
   - Address any feedback
   - Once approved, PR will be merged

## Plugin Best Practices

### Security
- Never commit secrets or API keys
- Use environment variables for sensitive data
- Validate all user inputs
- Review bash commands for security risks

### Performance
- Keep hooks lightweight
- Use progressive disclosure for skills
- Avoid unnecessary tool calls
- Cache when appropriate

### User Experience
- Provide clear error messages
- Include helpful examples
- Write descriptive documentation
- Follow principle of least surprise

### Maintainability
- Write self-documenting code
- Keep components focused
- Avoid premature optimization
- Refactor when needed

## Reporting Issues

### Bug Reports

Include:
- Claude Code version
- Plugin version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages (if any)
- Screenshots (if relevant)

### Feature Requests

Include:
- Clear description of feature
- Use case and motivation
- Proposed implementation (if any)
- Examples of similar features
- Potential challenges

## Questions?

- Check the [Claude Code documentation](https://code.claude.com/docs)
- Review existing issues and PRs
- Ask in discussions (if available)
- Contact maintainers

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Welcome newcomers
- Accept constructive criticism
- Focus on what's best for the community
- Show empathy

### Unacceptable Behavior

- Harassment or discrimination
- Trolling or insulting comments
- Personal or political attacks
- Publishing others' private information
- Unprofessional conduct

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

---

Thank you for contributing! 🎉
