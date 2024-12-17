# Trial Overview: Abdul Malek
## Expectations
- **Collaboration**: If you have questions, we encourage you to ask proactively! Our culture is quite collaborative, please be asking questions as there will be areas of the codebase that you don’t understand. 
- **The project**: We've picked a project that's deliberately open-ended, to reflect the type of work that we do--not sandboxed. It is nice to "finish" the project but as you know engineering projects don't tend to work this way. 
The goal of this project is to get a signal on your engineering ability in a product-oriented (not necessarily a trading) environment.
- **Check-ins**: At the end of each day, we’ll have a scheduled informal check-in to talk about your progress.

## Logistical Notes
- **The office**: Please keep the office key with you for the duration of the trial, it will give you access to the building. If you are leaving after the doorman, press the green button next to the lobby door to unlock the exit.
- **Working hours**: We usually get to the office around 10, feel free to come earlier/later after your first day. Also, the team sometimes works late, do not feel pressured to stay longer than 8 hours.
- **Meals**: Lunch and dinner is provided in-office. Help yourself to snacks and drinks!
- **Computer**: The laptop provided should stay at the office. Please do not take it home.

## Project: Iterated Blake3 Hashing
### Problem
In the SVM L2, one of the biggest challenges is efficiently computing a hash over all accounts in the state. You can think of an account as `Vec<u8>` where this vector can be up to 10MB. We want to compute this hash after every transaction, but we should avoid O(N) operations over the full vector. Ideally, we only rehash the sections of the vector that changed.

Our solution segments every account into chunks (pages) of size 4096 and builds a full binary Merkle tree out of the chunks (i.e. the chunks become the leaf nodes of the Merkle tree).

The current algorithm gathers all the modified pages in any given transaction, hashes the new page (leaf), and recomputes the root node of the full account Merkle tree (using the blake3 hash function to compute parent nodes).

Because the blake3 hash function uses an internal Merkle tree, we should be able to improve downstream developer ergonomics while improving the efficiency of our algorithm. 

The idea behind this project is to build a wrapper around blake3 that enables us to incrementally and efficiently recompute the root hash of an account given small changes to its data. This new account hash should have the following property:
```rust
account_hash(data) = blake3(data)
```

Note that the current construction looks like this for an account with 4 pages
```rust
CHUNK = 4096
leaf0 = blake3(data[:CHUNK])
leaf1 = blake3(data[CHUNK:2*CHUNK])
leaf2 = blake3(data[2*CHUNK:3*CHUNK])
leaf3 = blake3(data[3*CHUNK:4*CHUNK])
root_left = blake3([leaf0, leaf1])
root_right = blake3([leaf2, leaf3])
account_hash(data) = blake3([root_left, root_right])
```
​
### Solution
Build a custom implementation of Blake 3 that supports incremental leaf node updates. The incremental update should produce identical output to the library function (but should be much faster). We can keep an efficient implementation in the main sequencer node but a 1-time verifier can use a library function.

A binary Merkle tree implementation will be provided to you.

You will need to look through the implementation details of the Blake 3 algorithm and understand the topology of its internal state.

We'd like your solution to adapt the APIs of the provided source code and include comprehensive tests to demonstrate correctness.

### Resources
BLAKE3 Github: https://www.blake3.io  
Starter Code for merkle tree: https://gist.github.com/jarry-xiao/5d9896933f2414effe828d94d30f865f