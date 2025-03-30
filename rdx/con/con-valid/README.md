# con-valid

`con-valid` validates a transaction or a block against the respective state.

## Dependencies

For running the program, the `rdx` library must be installed and available in the system PATH.

## How to build

Run ``go build``

## How to run

- To validate transaction: ``./con-valid [--malicious] transaction <path to DB> <transaction_hash>``
- To validate \<path to DB\>/proposed_block.rdx block: ``./con-valid [--malicious] proposed-block <path to DB>``
