# Frozen Bash parity baseline

This directory preserves the pre-Go `scripts/` tree from commit
`d658db620836c4113e5a49326b5c69012c3e1f18`. The parity integration tests use
it as their immutable Bash oracle, so the test suite does not require Git
history to be available at runtime.
