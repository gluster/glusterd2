# Transaction framework

The GD2 transaction framework is used to execute/orchestrate distributed actions (transactions) over a Gluster trusted storage pool.
It is used to perform the various actions required by the different volume management and cluster management operations supported by GD2.

<!-- vim-markdown-toc GFM -->

* [Transaction](#transaction)
  * [Transaction step](#transaction-step)
* [Transaction engine](#transaction-engine)
  * [Creating and running a transaction.](#creating-and-running-a-transaction)
  * [Modify global data structures](#modify-global-data-structures)
  * [Synchronized step execution](#synchronized-step-execution)
  * [Cleaning up stale and failed transactions](#cleaning-up-stale-and-failed-transactions)
  * [Handling peer restart during transaction](#handling-peer-restart-during-transaction)
* [Examples](#examples)
  * [Volume create](#volume-create)
    * [Happy path: All peers alive throught out the transaction](#happy-path-all-peers-alive-throught-out-the-transaction)
    * [Fail path: initiator dies and comes back up within cleanup timeout](#fail-path-initiator-dies-and-comes-back-up-within-cleanup-timeout)
    * [Fail path: initiator dies and comes back up after first cleanup timeout](#fail-path-initiator-dies-and-comes-back-up-after-first-cleanup-timeout)
    * [Happy path: one executor dies and comes back up before cleanup/transaction timeout](#happy-path-one-executor-dies-and-comes-back-up-before-cleanuptransaction-timeout)
    * [Fail path: one executor dies and comes back up after cleanup/transaction timeout](#fail-path-one-executor-dies-and-comes-back-up-after-cleanuptransaction-timeout)
    * [Examples for more complex cases](#examples-for-more-complex-cases)
* [Terms](#terms)
  * [Data structures](#data-structures)
    * [Global data structures](#global-data-structures)
    * [Local data structures](#local-data-structures)
  * [Initiator](#initiator)
  * [Cleanup leader](#cleanup-leader)
  * [Locks](#locks)
    * [Cluster locks](#cluster-locks)
    * [Local locks](#local-locks)
  * [Stale transaction](#stale-transaction)
  * [Transaction namespaces](#transaction-namespaces)
    * [Pending transaction namespace](#pending-transaction-namespace)
    * [Transaction context namespace](#transaction-context-namespace)

<!-- vim-markdown-toc -->

## Transaction

A transaction is basically a collection of steps or actions to be performed in order.
A transaction object provides the framework with the following,

1. a list of peers that will be a part of the transaction
2. a set of transaction [steps](#transaction-step)

Given this information, the GD2 transaction framework will,

- verify if all the listed peers are online
- run each step on all of the nodes, before proceeding to the next step
- if a step fails, undo the changes done by the step and all previous steps.

The base transaction is basically free-form, allow users to create any order of steps.
This keeps it flexible and extensible to create complex transactions.

### Transaction step
A step is an action to be performed, most likely a function that needs to be run.
A step object provides the following information,

1. The function to be run
2. The list of peers the step should be run on.
3. An undo function that reverts any changes done by the step.

Each step can have its own list of peers, so that steps can be targeted to specific nodes and provide more flexibility.


## Transaction engine

The transaction engine executes the given transaction across the cluster.
The engine is designed to make use of etcd as the means of communication between peers.
The framework has to provide two important characteristics,

1. Each peer must be capable of independently and asynchronously execute a transaction that has been intitiated.
2. Each peer should be capable of independetly rollback/undo [stale transaction](#stale-transaction).

In addition the transaction engine should also provide,

1. A method to obtain [cluster wide locks](#cluster-locks) and [local locks](#local-locks),
   so that updates to [global](#global-data-structures) and [local](#local-data-structures) can be done safely.
2. The ability to synchronize transaction steps across the cluster when required.

The transaction engine is started on each peer in the cluster,
and keeps a watch on the [pending transaction namespace](#pending-transaction-namespace) for new transactions.
For each new incoming transactions the transaction engine does the following,

- Check if it peer is invovled in the transaction. If it is not, then do nothing.
- Fetch the list of steps for the transaction. For each step,
  - Check if transaction has been marked as failure
    - If transaction has failed,
      - rollback previously executed steps
      - mark yourself as having failed the transaction
      - end transaction
  - [Synchronize step](#synchronized-step-execution) if required
  - Check if step needs to be executed on peer
    - if not required, mark successfull progress and contine to next step
  - Execute the step
    - If step executes successfully,
      - mark progress in the [transaction namespace](#transaction-context-namespace)
      - continue to next step
    - If step fails,
      - rollback previously executed steps
      - mark yourself as having failed the transaction
      - end transaction
- After all steps have been executed,
  - mark yourself as having successfully completed the transaction
  - start a timeout timer and wait for transaction to be cleaned up.
    - If transaction is cleaned up
      - end transaction
    - If a timeout occurs,
        - rollback previously executed steps
        - mark yourself as having failed the transaction
        - end transaction

### Creating and running a transaction.

A transaction is initiated by an [initiator](#initiator).
The initiator is most likely the node that recives an incoming GD2 request.
The initiator does the following,

- Based on the incoming request, the initiator creates a transaction
  - If required the intiator takes any required cluster locks
    - If required the initiator can obtain locks before filling out the transaction steps and starting the transaction
- Add the created and filled transaction into the pending transaction namespace
- Start a timeout timer and watch for involved peers to mark transaction completion in transaction context
  - If all involved peers mark successful completion
    - Cleanup transaction
    - Respond back with result
  - If at least one peer marks failure
    - Mark transaction as having failed
    - Respond with error
  - If timeout occurs
    - Mark transaction as having failed
    - Respond with error

### Modify global data structures

[Global data structures](#global-data-structures) can only be updated under [cluster locks](#cluster-locks).
During a transaction it is required that,
- only the initiator modifies global data structures
- modification is only done as the last step of a transaction
- the transaction is synchronized before the modify step is executed

Modifications once done to global data structures cannot be rolled-back.

### Synchronized step execution

A synchornized step is executed only after all pervious steps have been completed successfully by all involved peers.
Step synchronization is required for steps that collate information, update global data structures or perform other similar operations.
The initiator can mark a step as synchronized when creating the transaction.

Step synchronization is performed by the engine for any synchornized step, even if the step would be executed on the peer.
The engine synchronizes as follows,
- Check if step needs synchronization
  - if not required
    - continue with step execution
  - if required
    - wait for previous step to marked as completed by all peers involved in transaction
      - if all peers mark completion
        - continue step execution
      - if any peer marks step failure,
        - mark yourself as having failed current step and return

Step synchronization is only done for forward execution of transactions, not for rollbacks.


### Cleaning up stale and failed transactions

A [leader](#cleanup-leader) is elected among the peers in the cluster to cleanup [stale transactions](#stale-transaction).
The leader periodically scans the pending transaction namespace for failed and stale transactions,
and cleans them up if rollback is completed by all peers involved in the transaction.

- After winning election or after hitting cleanup timer,
  - Fetch pending transactions
  - For each transaction,
    - if transaction is failed
      - ensure all peers have performed rollbacks (marked transaction as failure)
      - cleanup transaction from pending transactions
      - continue to next transaction
    - if transaction is stale (initiator down or transaction is active for longer than [transaction timeout](#transaction-timeout))
      - check if all peers have marked transaction as failure
        - if all peers have marked transaction as failure,
          - cleanup transaction from pending transactions
          - continue to next transaction
        - if not,
          - mark transaction as failure to trigger peers to perform rollbacks
          - continue to next transaction
  - restart cleanup timer

### Handling peer restart during transaction

If peer dies in the middle of transaction execution, and later restarts,
it will attempt to resume or rollback any transactions it was involved in.
This happens as follows,

- On peer startup, it scans the pending transaction namespace for transactions involving the peer
- For each such transaction,
  - Check if transaction has been marked as failure
    - if the transaction is marked as failure or you are transaction initiator
      - perform rollback from last completed step
      - mark yourself as having failed transaction
    - if not,
      - resume transaction execution from last completed step

Transactions cannot be safely resumed on initiators as any global locks it held will be lost when the peer died.


## Examples

The following assumptions are made.

- Cluster of size 3, with peers A, B and C.
- Peer A is always the initiator
- Peer B is the cleanup leader


### Volume create

Attempt to create a volume with bricks on all 3 peers.
Transaction created is as follows,
```
- Transaction: Create volume - vol1
  Initiator: A
  Nodes: A, B, C
  StartTime: T0
  GlobalLocks: Vol/vol1
  Steps:
  - Step: Check brick path
    Nodes: A B C
  - Step: Create brick xattrs
    Undo: Remove brick xattrs
    Nodes: A B C
  - Step: Create brickinfo and store brickinfo
    Undo: Remove stored brickinfo
    Nodes: A B C
  - Step: Create and store volinfo
    Sync: yes
    Node: A
```

#### Happy path: All peers alive throught out the transaction

|A (initiator)| A (engine)|B (engine)| C (engine)|B(cleanup)|
|---|---|---|---|---|
||Wait for new transactions|Wait for new transactions|Wait for new transactions|Start cleanup timer|
|Receive create request |||||
|Create transaction-1 and add to pending transactions|||||
|Wait for nodes to succeed or fail|||||
||New transaction-1|New transaction-1|New transaction-1||
||Execute steps 1-3, and mark as completed|Execute steps 1-3 and mark as completed|Execute steps 1-3 and mark as completed||
||Wait for all peers to complete step 3|Wait for all peers to complete step 3|Wait for all peers to complete step 3||
||Execute step 4|Skip step 4|Skip step 4||
||No more steps, mark self as succeeded transaction|No more steps, mark self as succeeded transaction|No more steps, mark self as succeeded transaction||
|All peers succeeded|||||
|Cleanup transaction-1|||||
|Send response|||||

#### Fail path: initiator dies and comes back up within cleanup timeout

|A (initiator)| A (engine)|B (engine)| C (engine)|B(cleanup)|
|---|---|---|---|---|
||Wait for new transactions|Wait for new transactions|Wait for new transactions|Start cleanup timer|
|Receive create request |||||
|Create transaction-1 and add to pending transactions|||||
|Wait for nodes to succeed or fail|||||
||New transaction-1|New transaction-1|New transaction-1||
|Peer dies||Execute steps 1-3 and mark as completed|Execute steps 1-3 and mark as completed||
|||Wait for all peers to complete step 3|Wait for all peers to complete step 3||
|Peer restarts|||||
||Check for pending transations||||
||Pending transaction-1||||
||Rollback transaction (as peer was initiator)||||
||Mark self as failed transaction-1||||
|||||Timer expires|
|||||Get pending stale transactions|
|||||Pending transaction-1 found|
|||||Mark transaction-1 as failed (not all peers have marked failed|
|||||Restart timer|
|||Transaction marked as failure|Transaction marked as failure||
|||Rollback transaction-1|Rollback transaction-1||
|||Mark self as failed transaction-1|Mark self as failed transaction-1||
|||||Timer expires|
|||||Get pending stale transactions|
|||||Pending transaction-1 found|
|||||All peers failed, delete transaction-1|
||||||

#### Fail path: initiator dies and comes back up after first cleanup timeout
|A (initiator)| A (engine)|B (engine)| C (engine)|B(cleanup)|
|---|---|---|---|---|
||Wait for new transactions|Wait for new transactions|Wait for new transactions|Start cleanup timer|
|Receive create request |||||
|Create transaction and add to pending transactions|||||
|Wait for nodes to succeed or fail|||||
||New transaction|New transaction|New transaction||
|Peer dies||Execute steps 1-3 and mark as completed|Execute steps 1-3 and mark as completed||
|||Wait for all peers to complete step 3|Wait for all peers to complete step 3||
|||||Timer expires|
|||||Get pending stale transactions|
|||||Pending transaction-1 found|
|||||Mark transaction-1 as failed (not all peers have marked failed|
|||||Restart timer|
|||Transaction marked as failure|Transaction marked as failure||
|||Rollback transaction-1|Rollback transaction-1||
|||Mark self as failed transaction-1|Mark self as failed transaction-1||
|Peer restarts|||||
||Check for pending transations||||
||Pending transaction-1||||
||Rollback transaction-1 (failed transaction)||||
||Mark self as failed transaction-1||||
|||||Timer expires|
|||||Get pending stale transactions|
|||||Pending transaction-1 found|
|||||All peers failed, delete transaction-1|
||||||

#### Happy path: one executor dies and comes back up before cleanup/transaction timeout

|A (initiator)| A (engine)|B (engine)| C (engine)|B(cleanup)|
|---|---|---|---|---|
||Wait for new transactions|Wait for new transactions|Wait for new transactions|Start cleanup timer|
|Receive create request |||||
|Create transaction and add to pending transactions|||||
|Wait for nodes to succeed or fail|||||
||New transaction|New transaction|New transaction||
||Execute steps 1-3 and mark as completed|Execute steps 1-3 and mark as completed|Peer dies||
||Wait for all peers to complete step 3|Wait for all peers to complete step 3|||
||||Peer restarts||
||||Check for pending transactions||
||||Get transaction-1||
||||Resume transaction-1||
||||Complete step-3||
||||Wait for all peers to complete step 3||
||Execute step 4|Skip step 4|Skip step 4||
||No more steps, mark self as succeeded transaction-1|No more steps, mark self as succeeded transaction-1|No more steps, mark self as succeeded transaction-1||
|All peers succeeded|||||
|Cleanup transaction-1|||||
|Send response|||||

#### Fail path: one executor dies and comes back up after cleanup/transaction timeout

|A (initiator)| A (engine)|B (engine)| C (engine)|B(cleanup)|
|---|---|---|---|---|
||Wait for new transactions|Wait for new transactions|Wait for new transactions|Start cleanup timer|
|Receive create request |||||
|Create transaction and add to pending transactions|||||
|Wait for nodes to succeed or fail|||||
||New transaction|New transaction|New transaction||
||Execute steps 1-3 and mark as completed|Execute steps 1-3 and mark as completed|Peer dies||
||Wait for all peers to complete step 3|Wait for all peers to complete step 3|||
|Transaction-1 timer expires|||||
|Mark transaction-1 as failed|||||
|Send error response|||||
||Transaction-1 marked failure|Transaction-1 marked failure|||
||Rollback|Rollback|||
||Mark self as failed transaction-1|Mark self as failed transaction-1|||
||||Peer restarts||
||||Check for pending transactions||
||||Get transaction-1||
||||Rollback transaction-1||
||||Mark self as failed transaction-1||
||||Wait for all peers to complete step 3||
|||||Timer expires|
|||||Get pending stale transactions|
|||||Pending transaction-1 found|
|||||All peers failed, delete transaction-1|

#### Examples for more complex cases
**TODO**

## Terms

### Data structures

#### Global data structures

Global data structures are the objects that span over multiple peers in the cluster. These include volumes, snapshots and the like. Updates to these data structures require that a [cluster lock](#global-locks) be obtained on them.

#### Local data structures

Local data structures are objects that are restricted to individual peers in the cluster. These include bricks, daemon processes etc. Updates to these data structures require that a [local lock](#local-locks) be obtained on them.

### Initiator

The intiator is the peer that initiates a transaction. The intiator prepares the list of transaction steps, adds them to the [pending transaction namespace](#pending-transaction-namespace), waits for the transaction to complete, and finally cleans-up the transaction from the new transaction namespace.

### Cleanup leader

The leader cleans-up any [stale transactions](#stale-transaction) from the [pending transaction namespace](#pending-transaction-namespace). The leader waits till the peers involved in the stale transaction have performed a rollback, before removing the transaction. Leaders are elected using etcd election mechanisms.

### Locks

#### Cluster locks

Locks taken to synchronize access to [global data structures](#global-data-structures). These locks will most likely be implemented as etcd locks, and are co-operative in nature.

#### Local locks

Locks taken to synchronize access to [local data structures](#local-data-structures). The locks will most likely be implemented as mutexes.

### Stale transaction

A stale transaction is a transaction where the transaction initiator is dies before the transaction completes, which results in the transaction never being cleaned up.

### Transaction namespaces

#### Pending transaction namespace

This is an etcd namespace, into which the [initiator](#initiator) adds new transactions. All peers keep a watch on this namespace for new transactions and execute transactions they are marked as being part of.

#### Transaction context namespace

Each individual transaction is provided with an etcd namespace, which is used to store/retrieve/share transaction specific contextual information when a transaction is being executed.
