/*
This module includes an implementation of a Binary Merkle Tree


The Binary Merkle Tree is 1-indexed and is constructed out of leaves equal to a power of two.
The tree is represented as a vector in which the root is at index 1, followed by its children, etc.
The leaves are in the second half of the vector.
The first node is a filler node that is a hash of &[0; PAGE_SIZE].
Indexing from 1 was chosen because the index math to find any node's parent, sibling, and whether the node is a left-node becomes simpler.

*/
use std::collections::VecDeque;
use std::iter::FromIterator;
use std::time::Instant;
mod blake3;

fn get_parent(left: &blake3::Output, right: &blake3::Output) -> blake3::Output {
    let parent_output =
        blake3::parent_output(left.chaining_value(), right.chaining_value(), blake3::IV, 0);
    parent_output
}

/// Binary merkle tree that is 1-indexed and is constructed out of leaves equal to a power of two.
/// If the number of leaves is not a power of two, add zero nodes until the number of leaves is a power of two.
#[derive(Debug, Clone)]
pub struct BinaryMerkleTree {
    pub tree: Vec<blake3::Output>,
}

impl BinaryMerkleTree {
    pub fn new_from_leaves(leaves: Vec<blake3::Output>) -> BinaryMerkleTree {
        // Initialize a zero vector with the correct number of nodes
        let number_of_leaves = leaves.len().next_power_of_two();
        let mut tree = Self::new_empty(number_of_leaves as u64);

        tree.create_tree_from_leaves(leaves);

        tree
    }

    pub fn root(&self) -> blake3::Output {
        self.tree[1]
    }

    pub fn num_leaves(&self) -> usize {
        self.tree.len() / 2
    }

    pub fn get_tree_length(&self) -> usize {
        self.tree.len() - 1 // Minus one because the tree is 1-indexed
    }

    pub fn new_empty(number_of_leaves: u64) -> Self {
        assert!(number_of_leaves.is_power_of_two());
        let tree: Vec<blake3::Output> =
            vec![blake3::Output::new([0; 16], 0); 2 * number_of_leaves as usize]; // We don't subtract one because the tree is 1-indexed
        BinaryMerkleTree { tree }
    }

    // The parent of a node is always at node_index / 2
    fn get_parent_index(index: usize) -> usize {
        index >> 1
    }

    fn create_tree_from_leaves(&mut self, leaves: Vec<blake3::Output>) {
        // Copy the leaves into the end of the tree
        let number_of_leaves = leaves.len();
        self.tree
            .splice(self.tree.capacity() - number_of_leaves.., leaves);
        // If there is only one leaf (plus the filler first node), the tree is simply that leaf
        if number_of_leaves == 1 {
            return;
        }

        // Build ancestors
        let leaf_start_index = self.get_tree_length() / 2 + 1;
        let leaves_with_indices = self.tree[leaf_start_index..]
            .iter()
            .copied()
            .zip(leaf_start_index..leaf_start_index + number_of_leaves);
        let mut hash_queue = VecDeque::from_iter(leaves_with_indices);
        while hash_queue.len() > 1 {
            let (left_child, left_index) = hash_queue.pop_front().unwrap();
            let (right_child, _right_index) = hash_queue.pop_front().unwrap(); // There should always be another node in the queue
            let parent_index = BinaryMerkleTree::get_parent_index(left_index);

            let parent_hash = get_parent(&left_child, &right_child);
            self.tree[parent_index] = parent_hash;
            hash_queue.push_back((parent_hash, parent_index));
        }
    }

    /// Update a leaf and propogate updates to all ancestors.
    /// Leaf index input is 0-indexed where the first leaf is index 0
    /// Leaf_index input should be 0-indexed where the first leaf would be entered as index 0
    pub fn update_leaf(&mut self, leaf_index: usize, leaf: blake3::Output) {
        let real_leaf_index = leaf_index + self.num_leaves();
        if self.tree[real_leaf_index].chaining_value() == leaf.chaining_value() {
            return;
        }
        self.tree[real_leaf_index] = leaf;

        let mut current_index = real_leaf_index;
        while current_index > 1 {
            // Update parent
            let parent_index = BinaryMerkleTree::get_parent_index(current_index);

            let (left_node_index, right_node_index) =
                self.get_left_and_right_node_indices_from_index(current_index);
            let left_node = &self.tree[left_node_index];
            let right_node = &self.tree[right_node_index];

            let parent_hash = get_parent(left_node, right_node);
            self.tree[parent_index] = parent_hash;
            current_index = parent_index;
        }
    }

    /// Bulk insert leaves and propogate hash updates to all ancestors.
    /// This method avoid updating shared parents if given two direct siblings to update.
    /// Leaf_index input should be 0-indexed where the first leaf would be entered as index 0
    pub fn bulk_insert_leaves<I, J>(
        &mut self,
        leaf_indices_iter: I,
        leaf_hashes_iter: J,
    ) -> Option<()>
    where
        I: Iterator<Item = usize>,
        J: Iterator<Item = blake3::Output>,
    {
        // Check if sorted
        let leaf_offset = self.num_leaves();
        let leaf_indices = leaf_indices_iter
            .map(|input_index| input_index + leaf_offset)
            .collect::<Vec<_>>();

        // In-line our own sort checker because Rust's is_sorted is not yet stable.
        fn is_sorted(leaf_indices: &[usize]) -> bool {
            (0..leaf_indices.len() - 1).all(|i| leaf_indices[i] < leaf_indices[i + 1])
        }
        if !is_sorted(&leaf_indices) {
            return None;
        }

        // Insert all leaf nodes
        for (leaf_index, updated_leaf_hash) in leaf_indices.iter().zip(leaf_hashes_iter) {
            self.tree[*leaf_index] = updated_leaf_hash;
        }

        // Update ancestors based on sorted leaf indices
        let mut update_queue = VecDeque::from(leaf_indices);
        while let Some(current_index) = update_queue.pop_front() {
            // Break if the root is reached
            if current_index == 1 {
                break;
            }

            // If the next ancestor to update is the sibling's, pop it from the queue
            // since it will have the same parent as the current node
            let sibling_index = BinaryMerkleTree::get_sibling_index(current_index);
            if let Some(&next_index) = update_queue.front() {
                if next_index == sibling_index {
                    update_queue.pop_front();
                }
            }

            let (left_node_index, right_node_index) =
                self.get_left_and_right_node_indices_from_index(current_index);
            let left_node = self.tree[left_node_index];
            let right_node = self.tree[right_node_index];

            let parent_index = BinaryMerkleTree::get_parent_index(current_index);
            let parent_hash = get_parent(&left_node, &right_node);
            self.tree[parent_index] = parent_hash;
            update_queue.push_back(parent_index);
        }

        Some(())
    }

    fn get_sibling_index(index: usize) -> usize {
        // Bit-wise XOR to get the sibling index
        // Example: Sibling of index 4(0b100) is 5(0b101) and sibling of index 5(0b101) is 4(0b100)
        index ^ 1
    }

    fn is_left(index: usize) -> bool {
        // All left-children have an even node index
        index % 2 == 0
    }

    /// Given an index of the current node, identify its direct sibling,
    /// identify which node is left, which is right, and return them.
    fn get_left_and_right_node_indices_from_index(&self, current_index: usize) -> (usize, usize) {
        let sibling_index = BinaryMerkleTree::get_sibling_index(current_index);

        // Use boolean indexing to avoid if statement branching
        let node_pair = [current_index, sibling_index]; // Stack allocation

        // If the sibling is the left child, is_left returns 1 and gets the sibling
        // If the sibling is the right child, is_left returns 0 and gets the node to update (the left child)
        let left_node_index = node_pair[BinaryMerkleTree::is_left(sibling_index) as usize];

        // If the node to update is the left child, is_left returns 1 and gets the sibling (the right child)
        // If the node to update is the right child, is_left returns 0 and gets the node to update
        let right_node_index = node_pair[BinaryMerkleTree::is_left(current_index) as usize];

        (left_node_index, right_node_index)
    }
}

fn main() {}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_update_leaf_correctness() {
        let exp_leaves: Vec<blake3::Output> = vec![
            blake3::Output::new(unsafe { std::mem::transmute([b'X'; 64]) }, 0),
            blake3::Output::new(unsafe { std::mem::transmute([b'B'; 64]) }, 1),
            blake3::Output::new(unsafe { std::mem::transmute([b'C'; 64]) }, 2),
            blake3::Output::new(unsafe { std::mem::transmute([b'S'; 64]) }, 3),
        ];
        let exp_tree = BinaryMerkleTree::new_from_leaves(exp_leaves);

        let act_leaves: Vec<blake3::Output> = vec![
            blake3::Output::new(unsafe { std::mem::transmute([b'A'; 64]) }, 0),
            blake3::Output::new(unsafe { std::mem::transmute([b'B'; 64]) }, 1),
            blake3::Output::new(unsafe { std::mem::transmute([b'C'; 64]) }, 2),
            blake3::Output::new(unsafe { std::mem::transmute([b'D'; 64]) }, 3),
        ];
        let mut act_tree = BinaryMerkleTree::new_from_leaves(act_leaves);

        let index0: usize = 0;
        act_tree.update_leaf(
            index0,
            blake3::Output::new(unsafe { std::mem::transmute([b'X'; 64]) }, index0 as u64),
        );
        let index3: usize = 3;
        act_tree.update_leaf(
            index3,
            blake3::Output::new(unsafe { std::mem::transmute([b'S'; 64]) }, index3 as u64),
        );

        let mut exp_out = [0u8; 32];
        let mut act_out = [0u8; 32];
        exp_tree.root().root_output_bytes(&mut exp_out);
        act_tree.root().root_output_bytes(&mut act_out);

        assert_eq!(exp_out, act_out);
    }

    #[test]
    fn test_blake3_correctness() {
        let exp_leaves = &[[b'A'; 64], [b'B'; 64], [b'C'; 64], [b'D'; 64]].concat();

        let mut b3hasher = blake3::Hasher::new();
        b3hasher.update(&exp_leaves);
        let mut exp_hash = [0u8; 32];
        b3hasher.finalize(&mut exp_hash);

        let act_leaves: Vec<blake3::Output> = vec![
            blake3::Output::new(unsafe { std::mem::transmute([b'A'; 64]) }, 0),
            blake3::Output::new(unsafe { std::mem::transmute([b'B'; 64]) }, 1),
            blake3::Output::new(unsafe { std::mem::transmute([b'C'; 64]) }, 2),
            blake3::Output::new(unsafe { std::mem::transmute([b'D'; 64]) }, 3),
        ];
        let act_tree = BinaryMerkleTree::new_from_leaves(act_leaves);

        let mut act = [0u8; 32];
        act_tree.root().root_output_bytes(&mut act);
        assert_eq!(exp_hash, act);
    }

    // #[test]
    // fn test_bulk_update_performance() {
    //     let num_updates = 10000;
    //     let leaves: Vec<blake3::Output> = (0..num_updates)
    //         .map(|i| blake3::Output::new(unsafe { std::mem::transmute([i as u8; 64]) }, i as u64))
    //         .collect();

    //     // Measure time for bulk hashing using blake3 hasher
    //     let start = Instant::now();
    //     let mut b3hasher = blake3::Hasher::new();
    //     for leaf in &leaves {
    //         let new: [u8; 32] = unsafe { std::mem::transmute( leaf.chaining_value() ) };
    //         b3hasher.update(&new);
    //     }
    //     let mut bulk_hash = [0u8; 32];
    //     b3hasher.finalize(&mut bulk_hash);
    //     let bulk_duration = start.elapsed();

    //     // Measure time for incremental updates using BinaryMerkleTree
    //     let mut tree = BinaryMerkleTree::new_from_leaves(leaves.clone());
    //     let start = Instant::now();
    //     for (i, leaf) in leaves.iter().enumerate() {
    //         tree.update_leaf(i, *leaf);
    //     }
    //     let incremental_duration = start.elapsed();

    //     assert!(incremental_duration < bulk_duration);

    //     // Ensure the root hash is the same
    //     let mut tree_root_hash = [0u8; 32];
    //     tree.root().root_output_bytes(&mut tree_root_hash);
    //     assert_eq!(bulk_hash, tree_root_hash);
    // }
}
