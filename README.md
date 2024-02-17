# word-of-wisdom
Design and implement “Word of Wisdom” tcp server.
• TCP server should be protected from DDOS attacks with the Prof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
• The choice of the POW algorithm should be explained.
• After Prof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.
• Docker file should be provided both for the server and for the client that solves the POW challenge

## How to use
You will need go 1.21+ installed to build binaries or run tests.

```
# Run both server and client in docker
make run-apps
```

## POW algorithm
As the PoW algorithm, a simplified analog of hashcash was chosen, in which we attempt to find a hash from the original challenge data and our nonce. 
This hash must meet a certain condition, namely having the required number of zeros at the beginning.

This algorithm was chosen because it is relatively simple to implement, widely used, and applied in Bitcoin, and it allows for dynamic adjustment of difficulty.
