# Contributing to Bolt

Thank you for your interest in contributing to Bolt! 🚀

## 🤝 **How to Contribute**

### **Types of Contributions**

- 🐛 **Bug Reports**: Report issues you encounter
- 💡 **Feature Requests**: Suggest new features
- 📝 **Documentation**: Improve docs and examples
- 🔧 **Code Contributions**: Submit pull requests
- 🧪 **Testing**: Help test and validate features

## 🚀 **Getting Started**

### **Prerequisites**

- Go 1.21 or higher
- OpenTofu installed
- Git

### **Development Setup**

```bash
# Clone the repository
git clone https://github.com/rioprayogo/bolt.git
cd bolt

# Install dependencies
go mod download

# Build the project
go build -o bold .

# Run tests
go test ./...
```

## 📋 **Contribution Guidelines**

### **Code Style**

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Write unit tests for new features

### **Commit Messages**

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(compiler): add Azure provider support
fix(parser): handle missing required fields
docs(readme): update installation instructions
```

### **Pull Request Process**

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes
4. **Test** your changes thoroughly
5. **Commit** with conventional commit format
6. **Push** to your fork
7. **Create** a Pull Request

### **PR Requirements**

- ✅ **Tests pass** - All existing and new tests must pass
- ✅ **Code coverage** - Maintain or improve test coverage
- ✅ **Documentation** - Update docs if needed
- ✅ **Examples** - Add examples for new features
- ✅ **Backward compatibility** - Don't break existing functionality

## 🧪 **Testing**

### **Run All Tests**
```bash
go test ./...
```

### **Run Specific Tests**
```bash
go test ./pkg/parser/...
go test -v ./pkg/compiler/...
```

### **Test Coverage**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📚 **Documentation**

### **Code Documentation**
- Add comments for exported functions
- Include usage examples
- Document complex algorithms
- Update README for new features

### **YAML Examples**
- Add examples for new providers
- Include both local and production configs
- Document all available parameters
- Provide real-world use cases

## 🐛 **Bug Reports**

### **Before Reporting**
- Check existing issues
- Try the latest version
- Reproduce the issue
- Gather relevant information

### **Bug Report Template**
```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Use YAML config '...'
2. Run command '....'
3. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- OS: [e.g. macOS, Linux, Windows]
- Bolt version: [e.g. 1.0.0]
- OpenTofu version: [e.g. 1.6.0]
- Cloud provider: [e.g. AWS, Azure, GCP]

**Additional context**
Add any other context about the problem.
```

## 💡 **Feature Requests**

### **Feature Request Template**
```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
A clear description of any alternative solutions.

**Additional context**
Add any other context or screenshots.
```

## 🏆 **Recognition**

Contributors will be:
- Listed in our [CONTRIBUTORS.md](CONTRIBUTORS.md) file
- Acknowledged in release notes
- Given credit for their contributions

## 📞 **Need Help?**

- 📖 **Documentation**: Check the [README](README.md)
- 💬 **Discussions**: Use GitHub Discussions
- 🐛 **Issues**: Create an issue for bugs
- 📧 **Email**: Contact us directly

Thank you for contributing to Bolt! 🎉 