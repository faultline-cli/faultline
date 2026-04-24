# Source Adapters

Faultline supports ingestion from multiple public source adapters. When building a mixed intake batch, prefer diversity across adapters to ensure broad coverage.

## Available Adapters

- **`github-issue`** - GitHub issues with failure reports and log snippets
- **`gitlab-issue`** - GitLab issues with failure context
- **`stackexchange-question`** - StackExchange questions (Stack Overflow, DevOps, ServerFault) with reproducible failure examples
- **`discourse-topic`** - Discourse forum threads with technical discussions and failure logs
- **`reddit-post`** - Reddit posts from technical subreddits (r/devops, r/golang, r/node, r/docker, etc.)

## Selection Priority

1. Prefer adapters with direct machine-produced log output over community discussions
2. Prioritize sources with full environment context and exact tool versions
3. Balance selection across adapters to avoid over-representation of any single source type
4. Bias toward underrepresented adapters when current corpus statistics show gaps