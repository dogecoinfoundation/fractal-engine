# Scratch

 
OK imagine this: 10,000 FE nodes all in a 100 x 100 grid, connected to the ones beside them,
a block arrives with an FE mint,
every node asks a peer for the metadata,
but no-one has it,
so they start the timer?
then, the original Minting node gossips out the metadata,
and it starts rippling across the grid,
some nodes might be in the middle of asking a peer for it,
when they receive it from dogenet,
so it needs to be resilient to that,
they might also receive it from dogenet AND get a reply from a peer with the metadata, so they receive it twice.
timer is still useful as a fallback in case you completely miss it

L1 -> N