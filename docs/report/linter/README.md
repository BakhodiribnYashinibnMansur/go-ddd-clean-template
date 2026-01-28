# Golangci-lint Report Configuration

## Overview
This project is configured to automatically generate comprehensive linting reports in multiple formats whenever `golangci-lint run` is executed.

## Report Location
All linting reports are automatically saved to:
```
docs/report/linter/
```

## Generated Report Formats

### 1. **report.txt** (Tab Format)
- Human-readable text format
- Tab-separated values
- Easy to parse with scripts
- **Size**: ~28 KB

### 2. **report.json** (JSON Format)
- Structured JSON data
- Complete issue details
- Perfect for programmatic analysis
- Used by admin panel for interactive display
- **Size**: ~76 KB

### 3. **report.html** (HTML Format)
- Interactive web interface
- Built with React and Bulma CSS
- Syntax highlighting
- Can be opened directly in browser
- **Size**: ~49 KB

### 4. **report.xml** (Checkstyle Format)
- XML format compatible with Checkstyle
- Integrates with CI/CD tools
- **Size**: ~37 KB

### 5. **codeclimate.json** (Code Climate Format)
- Code Climate compatible format
- For code quality platforms
- **Size**: ~61 KB

### 6. **junit.xml** (JUnit Format)
- JUnit XML test report format
- CI/CD integration
- **Size**: ~109 KB

### 7. **sarif.json** (SARIF Format)
- Static Analysis Results Interchange Format
- GitHub Advanced Security compatible
- **Size**: ~63 KB

### 8. **teamcity.txt** (TeamCity Format)
- TeamCity service messages format
- TeamCity CI integration
- **Size**: ~43 KB

## Usage

### Command Line
Simply run:
```bash
golangci-lint run
```

All reports will be automatically generated in `docs/report/linter/` directory.

### Admin Panel
1. Navigate to Admin Panel → Code Linter
2. Click "Run Linter" button
3. View interactive results with:
   - Search functionality
   - Accordion UI by linter type
   - Issue details with source code
   - Download buttons for all formats

## Configuration
The report generation is configured in `.golangci.yml`:

```yaml
output:
  formats:
    - format: colored-line-number
      path: stdout
    - format: json
      path: docs/report/linter/report.json
    - format: tab
      path: docs/report/linter/report.txt
    - format: html
      path: docs/report/linter/report.html
    - format: checkstyle
      path: docs/report/linter/report.xml
    - format: code-climate
      path: docs/report/linter/codeclimate.json
    - format: junit-xml
      path: docs/report/linter/junit.xml
    - format: teamcity
      path: docs/report/linter/teamcity.txt
    - format: sarif
      path: docs/report/linter/sarif.json
```

## Features

### Admin Panel Integration
- **JSON Parsing**: Automatically parses JSON report
- **Accordion UI**: Issues grouped by linter type
- **Search**: Real-time search across all issues
- **Sorting**: Linters sorted by issue count
- **Details**: File path, line number, column, description, source code
- **Download**: All formats available for download

### CI/CD Integration
All formats are ready for integration with:
- GitHub Actions (SARIF)
- Jenkins (JUnit, Checkstyle)
- TeamCity (TeamCity format)
- Code Climate (Code Climate format)
- SonarQube (Checkstyle, JUnit)

## Statistics
Based on current codebase analysis:
- **Total Issues**: 225
- **Active Linters**: 1 (typecheck)
- **Files Affected**: Multiple test files
- **Report Generation Time**: ~1-2 minutes

## Notes
- Reports are regenerated on each linter run
- Old reports are overwritten
- All formats are generated simultaneously
- No manual intervention needed
- Reports are served via `/docs/report/linter/` endpoint
