**Published on 3. July 2020**

Pirsch is a great success so far. At least that is what I would call it from looking at the [traffic](https://marvinblum.de/tracking) on my website and the stars on [GitHub](https://github.com/emvi/pirsch). A few people on [Hacker News](https://news.ycombinator.com/item?id=23668212) pointed out some details you should know in case you're using it.

> You can find a detailed article about server side tracking in Go [here](https://marvinblum.de/blog/server-side-tracking-without-cookies-in-go-OxdzmGZ1Bl).

Legal Stuff
-----------

While it is not possible to tell _who_ visited your website, it's still a good idea to mention tracking on your terms and conditions page (or whatever you call it). I'm not a lawyer, but the GDPR covers tracking methods that don't use cookies. You won't need a cookie banner (consent) as far as I can tell, because Pirsch does not collect personal data, which is one of the main goals I wanted to achieve, but as I said, I'm not a lawyer.

Fingerprinting
--------------

Another point that came up is how Pirsch generates fingerprints. The method is fine, it's just that it had one issue: the algorithm is open source and there is no randomness. Let me explain: if someone gets access to the fingerprints you generated, he could theoretically generate identical fingerprints for visitors to other websites, and therefore tell which websites a user visited by comparing them. I fixed this issue in release v1.1.0 by adding a salt you define. It should be set to something no one can guess and be treated like a password.

Filtering Bots
--------------

Additionally to the change above I extended the bot keyword list. It now includes everything that should not occur inside the User-Agent header. It now contains 365 entries, which should be enough to filter the most unwanted traffic.

Conclusion
----------

Thank you, everyone, for all the helpful feedback! I would love to hear if you're using it and how well it works for you. Just send me a mail using the button (Contact me) at the top.
