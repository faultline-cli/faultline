# Local Staging Queue

`faultline fixtures ingest` writes raw public-source candidates into this directory for local review.

Do not commit staged fixture YAML files directly. Sanitize sensitive fields first, then promote accepted cases into `fixtures/real/`.