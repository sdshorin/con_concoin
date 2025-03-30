# con-valid

`con-valid` validates a transaction or a block against the respective state.

## How to build

Run ``go build``

## How to run

- To validate transaction: ``./con-valid [--malicious] transaction <path to DB> <transaction_hash>``
- To validate \<path to DB\>/proposed_block block: ``./con-valid [--malicious] proposed-block <path to DB>``