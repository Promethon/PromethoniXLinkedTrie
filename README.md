# PromethoniXLinkedTrie

**Promethon:** New way to look at the next generation of Internet!

Ethereum uses a world state to hold states of each account.
But the size of this world state is growing rapidly and there hasn't been any efficient solution for shrinking it down.
World state is implemented by using [Merkle Patricia Trie](https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/) but today, we introduce **PromethoniXLinkedTrie**.
PromethoniXTrie acts as a bridge between [Merkle Patricia Trie](https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/) and [LinkedList]([https://en.wikipedia.org/wiki/Red%E2%80%93black_tree](https://en.wikipedia.org/wiki/Linked_list)).
The main idea is to **keep an extra value to validate** that the account should still exist on the network or not.
But due to the large volume of data, the challenge that has always existed is how to perform this validation in a short time and remove invalid accounts.

If we make a doubly linked list with the trie nodes and keep the linked list always sorted by the last modified block number, we can always find the items that we want to delete in Big O time of **O(1)**.

To be honest, this question can have many different answers, each with pros and cons. But what we all know is that eventually we want to reduce the volume of data, so we can't store all the data all the time.

- The main idea of this repo: we can consider the "extra value" as the block time of the last change or use in the contract. For example, if a year has passed since the last change, we can delete that account.
- Another idea is to charge a fee per certain amount of volume. Whenever this amount is used up and never charged again, delete that account. [Read More & Implementation](https://github.com/Promethon/PromethoniXTrie)

The **PromethoniXTrie** project implements the basis of the first idea. It creates a Merkle Patricia Trie that every value in the trie is a node of Linked List and establishes the desired connection between them and then tests the performance of the **Promethon** idea.

Our initial tests have been positive and promising. We are currently implementing this idea on the [go-ethereum](https://github.com/ethereum/go-ethereum) source code and we hope to announce its completion soon!

proposal for this aproach can be found [here](PromethoniXLinked.pdf)
