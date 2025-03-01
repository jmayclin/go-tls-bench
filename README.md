# Go TLS Bench
`go test -bench=. -v`

| Cert       | graviton 2 | graviton 3 |
| ---------  | ---------- | ------- |
| rsa2048    |  6.27 ms   | 2.29 ms |
| ecdsa256   |  2.09 ms   | 0.75 ms |
| ecdsa384   |  9.12 ms   | 3.64 ms |
| resumption |  1.05 ms   | 0.36 ms |

## Parameters
The cert chain is 3 long, consisting of a trusted CA, intermediate, and leaf cert. Handshake either use server auth (client verifies server cert) or resumption, as indicated by the table columns.

`go version go1.22.2 linux/arm64`

## Implementation
Disclaimer: I am very much not a Go programmer, this was stapled together from snippets that I found over the internet. It is largely drawing from similar work that we've done in s2n-tls.

The benchmark spawns two go-routines to drive the client and server, and the dummy connection mocks out network IO, using `chan byte[]` instead of the full TCP stack. I would have preferred to just use a single "thread", but Go seems to be committed to "blocking" IO (where the go-routine is blocked) which pushed me down the multiple go-routine route.

I have done some *horrendous* things to get my shutdown logic to work with the blocking-io model. I don't think it should cause any performance concerns, but wow it's pretty gnarly.


## Other Observations & Questions
- What's with the capitalized methods? e.g. `conn.Handshake()`. I do not understand that convention or what it is telling me.
- Wow the [Rust book](https://doc.rust-lang.org/book/) is so nice. I kept wanting to reach for something like that, but didn't really find a similar `Go` resource 😞. Presumably it exists, but I just wasn't able to find it with my low effort googling. (hehe, ironic)
- I don't understand Go's typing. This is probably partly my fault for just stapling together some GPT'd code, but the `net.Conn` dummy implementation makes me deeply uncomfortable. I think this is "duck typing"? Actually, just googled and maybe `net.Conn` is an interface and not a type? And I just sort of yeet the appropriate methods into existence without any other declaration, and that's good enough?
- The read polling behavior is confusing. When I made my first dummy conn impl based on the `byte.Buffer` types, it was a "non-blocking" implementation which lead to the method being called in a hot loop? Why did it stop calling? When I had removed the println, then things seemed to hang? One part of me wants to investigate this more deeply, the other part of me just doesn't want to touch anymore go.
- I don't like `nil`. Java did this, and I thought we all agreed that `null` was stupid and we were going to quit doing it.
- I miss my `?` operator 😭
- I was very confused by the module system. I just wanted a chunk of info like [this](https://doc.rust-lang.org/book/ch07-00-managing-growing-projects-with-packages-crates-and-modules.html), but didn't easily find it.
- I don't like the magic test detection, which seems to be based on method name/file name?

Side Note: I wonder how many of these were related to my usage of ChatGPT. I generally restrict ChatGPT usage to "interactive documentation search companion" rather than actual author of anything. 
1. Maybe I struggled with the language concepts because ChatGPT allowed me to leapfrog the basic learning process, causing me to stumble into confusing concepts without any of the prerequisites. Kinda like if a video game has really bad level scaling dynamics.
2. ChatGPT won't ever tell you "that's a stupid idea, do this instead", or at least I don't think to explicitly prompt it for that. I think if I had talked to an engineer familiar with go about my "benchmark using ring buffer/shared memory" they could have told me from the start that that was a moderately stupid idea and I should use channels instead.

### Graviton 3 Raw
```
goos: linux
goarch: arm64
pkg: tls-bench
BenchmarkServerAuthRsa2048
BenchmarkServerAuthRsa2048-16                524           2286291 ns/op
BenchmarkServerAuthEcdsa256
BenchmarkServerAuthEcdsa256-16              1576            749511 ns/op
BenchmarkServerAuthEcdsa384
BenchmarkServerAuthEcdsa384-16               327           3642075 ns/op
BenchmarkResumption
BenchmarkResumption-16                      3172            360598 ns/op
```

### Graviton 2 Raw
```
goos: linux
goarch: arm64
pkg: tls-bench
BenchmarkServerAuthRsa2048
BenchmarkServerAuthRsa2048-2    	     190	   6272629 ns/op
BenchmarkServerAuthEcdsa256
BenchmarkServerAuthEcdsa256-2   	     576	   2091612 ns/op
BenchmarkServerAuthEcdsa384
BenchmarkServerAuthEcdsa384-2   	     130	   9120546 ns/op
BenchmarkResumption
BenchmarkResumption-2           	    1083	   1052870 ns/op
```
