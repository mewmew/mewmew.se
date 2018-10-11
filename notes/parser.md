# Parser Techniques

For the difference between LR and LL, see http://blog.reverberate.org/2013/07/ll-and-lr-parsing-demystified.html. In short:

> LL parser outputs a pre-order traversal of the parse tree and an LR parser outputs a post-order traversal.

## LR

### LALR(1)

**pro**

- small state space
    + since states are merged

**con**

- leads to reduce/reduce conflicts
    + since states are merged
- leads to shift/reduce conflicts
    + when 1 lookahead token is not enough to disambiguate

### LR(1)

**pro**

- does not suffer from reduce/reduce conflicts

**con**

- huge state space
- leads to shift/reduce conflicts
    + when 1 lookahead token is not enough to disambiguate

## GLR(1)

**pro**

- does not suffer from reduce/reduce conflicts
- does not suffer from shift/reduce conflicts

**con**

- huge state space
- has to maintain a parse forest rather than a parse tree

## LL

TODO

## PEG

Terminology: a *packrat* parser is essentially a PEG parser utilizing dynamic programming.

**pro**

**con**

- difficult to write correct grammars
    + since choices are ordered (if the first alternative succeeds, succeding alternatives are ignored), *a* parse tree is deterministically chosen from the set of possible parse trees; but quite likely not the one you want or intend.

## Parser combinator

TODO
