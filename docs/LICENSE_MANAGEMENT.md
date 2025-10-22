# License Header Management

This project uses [Apache SkyWalking License-Eye](https://github.com/apache/skywalking-license-eye) to manage license headers in all source files.

## Installation

### Option 1: Docker (Recommended)

```bash
docker pull apache/skywalking-license-eye
```

### Option 2: Go Install

```bash
go install github.com/apache/skywalking-license-eye/cmd/license-eye@latest
```

### Option 3: Download Binary

Download from [GitHub Releases](https://github.com/apache/skywalking-license-eye/releases)

## Usage

### Check License Headers

Check if all files have proper license headers:

```bash
# Using Docker
docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-license-eye header check

# Using local binary
license-eye header check
```

### Fix License Headers

Automatically add missing license headers:

```bash
# Using Docker
docker run -it --rm -v $(pwd):/github/workspace apache/skywalking-license-eye header fix

# Using local binary
license-eye header fix
```

## Configuration

The configuration is in `.licenserc.yaml`:

```yaml
header:
  license:
    spdx-id: Apache-2.0
    copyright-owner: OneMoney
    copyright-year: 2025

  paths:
    - "**/*.go"
    - "**/*.md"

  paths-ignore:
    - "vendor/**"
    - "**/*.mod"
    - "**/*.sum"
    - ".git/**"
    - ".idea/**"

  comment: on-failure

  language:
    Go:
      extensions:
        - ".go"
      comment_style_id: DoubleSlash
```

## License Header Format

### For Go Files (.go)

```go
// Copyright 2025 OneMoney
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main
```

### For Markdown Files (.md)

```markdown
<!--
 Copyright 2025 OneMoney

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# Document Title
```

## CI/CD Integration

### GitHub Actions

Create `.github/workflows/license.yml`:

```yaml
name: License Check

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  license:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check License Header
        uses: apache/skywalking-eyes@main
        with:
          config: .licenserc.yaml
```

### GitLab CI

Add to `.gitlab-ci.yml`:

```yaml
license-check:
  stage: test
  image: apache/skywalking-license-eye
  script:
    - license-eye header check
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Checking license headers..."
license-eye header check

if [ $? -ne 0 ]; then
    echo "License header check failed. Run 'license-eye header fix' to fix."
    exit 1
fi
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Makefile Integration

Add to `Makefile`:

```makefile
.PHONY: license-check
license-check:
	@echo "Checking license headers..."
	@docker run -it --rm -v $(PWD):/github/workspace apache/skywalking-license-eye header check

.PHONY: license-fix
license-fix:
	@echo "Fixing license headers..."
	@docker run -it --rm -v $(PWD):/github/workspace apache/skywalking-license-eye header fix
```

Usage:
```bash
make license-check
make license-fix
```

## Quick Reference

```bash
# Check all files
license-eye header check

# Fix all files
license-eye header fix

# Check specific directory
license-eye -c .licenserc.yaml header check

# Dry-run (see what would be changed)
license-eye header fix --dry-run
```

## Troubleshooting

### Issue: License headers not detected

**Solution**: Check if file extensions are configured in `.licenserc.yaml`

### Issue: Some files should be ignored

**Solution**: Add patterns to `paths-ignore` in `.licenserc.yaml`

### Issue: Wrong copyright year

**Solution**: Update `copyright-year` in `.licenserc.yaml` and run `license-eye header fix`

### Issue: Custom license header needed

**Solution**: Create a custom header template file and reference it in `.licenserc.yaml`:

```yaml
header:
  license:
    spdx-id: Apache-2.0
    copyright-owner: OneMoney
    copyright-year: 2025
    content: |
      Copyright 2025 OneMoney

      Licensed under the Apache License, Version 2.0...
```

## Best Practices

1. **Run before committing**: Always check license headers before committing
2. **Use in CI/CD**: Add license header check to your CI pipeline
3. **Keep configuration updated**: Update copyright year annually
4. **Document exceptions**: If some files shouldn't have headers, document why in `paths-ignore`

## Resources

- [License-Eye Documentation](https://github.com/apache/skywalking-license-eye)
- [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)
- [SPDX License List](https://spdx.org/licenses/)
