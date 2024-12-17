/*
This module includes an implementation of a Binary Merkle Tree


The Binary Merkle Tree is 1-indexed and is constructed out of leaves equal to a power of two.
The tree is represented as a vector in which the root is at index 1, followed by its children, etc.
The leaves are in the second half of the vector.
The first node is a filler node that is a hash of &[0; PAGE_SIZE].
Indexing from 1 was chosen because the index math to find any node's parent, sibling, and whether the node is a left-node becomes simpler.

*/
use std::collections::VecDeque;
use std::hash::{DefaultHasher, Hasher};
use std::iter::FromIterator;
use blake3::Hasher as Blake3Hasher;

#[derive(Debug, Clone, Copy)]
pub struct Hash([u8; 32]);

impl Hash {
    fn new(src: &[u8; 32]) -> Self {
        Self(*src)
    }

    pub fn to_bytes(&self) -> [u8; 32] {
        self.0
    }
}

// Toy hash function
fn hashv(bytes: &[&[u8]]) -> Hash {
    let mut hasher = DefaultHasher::new();
    for val in bytes {
        hasher.write(val);
    }
    let x1 = hasher.finish().to_le_bytes();
    hasher.write(&x1);
    let x2 = hasher.finish().to_le_bytes();
    hasher.write(&x2);
    let x3 = hasher.finish().to_le_bytes();
    hasher.write(&x3);
    let x4 = hasher.finish().to_le_bytes();

    let res = [x1, x2, x3, x4].concat();

    let mut hash = [0u8; 32];
    hash.copy_from_slice(&res);
    Hash(hash)
}

/// Binary merkle tree that is 1-indexed and is constructed out of leaves equal to a power of two.
/// If the number of leaves is not a power of two, add zero nodes until the number of leaves is a power of two.
#[derive(Debug, Clone)]
pub struct BinaryMerkleTree {
    pub tree: Vec<Hash>,
}

impl BinaryMerkleTree {
    pub fn new_from_leaves(leaves: Vec<Hash>) -> BinaryMerkleTree {
        // Initialize a zero vector with the correct number of nodes
        let number_of_leaves = leaves.len().next_power_of_two();
        let mut tree = Self::new_empty(number_of_leaves as u64);

        tree.create_tree_from_leaves(leaves);

        tree
    }

    pub fn root(&self) -> Hash {
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
        let tree: Vec<Hash> = vec![Hash::new(&[0; 32]); 2 * number_of_leaves as usize]; // We don't subtract one because the tree is 1-indexed
        BinaryMerkleTree { tree }
    }

    // The parent of a node is always at node_index / 2
    fn get_parent_index(index: usize) -> usize {
        index >> 1
    }

    fn create_tree_from_leaves(&mut self, leaves: Vec<Hash>) {
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
            let parent_hash = hashv(&[&left_child.to_bytes(), &right_child.to_bytes()]);
            let parent_index = BinaryMerkleTree::get_parent_index(left_index);
            self.tree[parent_index] = parent_hash;
            hash_queue.push_back((parent_hash, parent_index));
        }
    }

    /// Insert a leaf and propogate updates to all ancestors.
    /// Leaf index input is 0-indexed where the first leaf is index 0
    /// Leaf_index input should be 0-indexed where the first leaf would be entered as index 0
    pub fn insert_leaf(&mut self, leaf_index: usize, leaf_hash: Hash) {
        let real_leaf_index = leaf_index + self.num_leaves();
        self.tree[real_leaf_index] = leaf_hash;

        let mut current_index = real_leaf_index;
        while current_index > 1 {
            // Update parent
            let parent_index = BinaryMerkleTree::get_parent_index(current_index);

            let (left_node_index, right_node_index) =
                self.get_left_and_right_node_indices_from_index(current_index);
            let left_node = &self.tree[left_node_index];
            let right_node = &self.tree[right_node_index];

            let parent_hash = hashv(&[&left_node.to_bytes(), &right_node.to_bytes()]);
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
        J: Iterator<Item = Hash>,
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

            let parent_hash = hashv(&[&left_node.to_bytes(), &right_node.to_bytes()]);
            let parent_index = BinaryMerkleTree::get_parent_index(current_index);
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

fn main() {
    let leaves = vec![
        Hash([b'A'; 32]),
        Hash([b'B'; 32]),
        Hash([b'C'; 32]),
        Hash([b'D'; 32]),
    ];
    let mut tree = BinaryMerkleTree::new_from_leaves(leaves);
    println!("{:?}", tree.root());

    tree.insert_leaf(0, Hash([b'E'; 32]));

    println!("{:?}", tree.root());
}
