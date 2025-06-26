/// Module: bridgemock
module bridgemock::bridgemock;

use sui::table::{Self, Table};


public struct Batch has copy, drop, store {
    nonce: u64,
    block_number: u64,
}

public struct Safe has key {
		id: UID,
        batches_count: u64,
		batches: Table<u64, Batch>,
}

fun init(ctx: &mut TxContext) {
    let safe = Safe {
            id: object::new(ctx),
            batches_count: 0,
            batches: table::new(ctx),
    };

    transfer::share_object(safe);
}

public entry fun add_batch(safe: &mut Safe, batch_nonce: u64, block_number: u64) {
    let nonce = safe.batches_count + 1;
    table::add(&mut safe.batches, safe.batches_count, Batch {
        nonce: batch_nonce,
        block_number
    });

    safe.batches_count = nonce;
}

public fun get_batch(safe: &Safe, nonce: u64): Batch {
    let batch = *table::borrow(&safe.batches, nonce - 1);
    batch
}

