**Published on 28. August 2020**

You can find a lot of articles about how to prevent deadlocks in Go, but most of them focus on concurrency patterns and synchronization tools like mutexes. While it is important to know some techniques to prevent them, a trap you can stumble across more easily without noticing, are database transaction deadlocks.

A transaction deadlock can occur when you start one or more transactions and run queries outside of transactions while they are still active. If you run too many transactions and queries at the same time, you might run out of database connections in the connection pool. Here is a simple example of that:

```
// We will ignore errors for this example,
// you should always check them of course.

tx, _ := db.Begin()
tx.Exec(`INSERT INTO "foo" ("a", "b") VALUES (4, 2)`)

// ...

db.Query(`SELECT * FROM "foo" WHERE "a" = $1 AND "b" = $2`, 4, 2)

// DEADLOCK

tx.Commit()
```

In this example, we create a new transaction and insert something to the database. Later on, we try to query the same result from the database. That the inserted row has not been committed yet, is not the actual issue, as you would receive no result in that case. The real issue here is that if you run this code concurrently, you might run out of connections. How many connections are opened to the database can be configured. As soon as your code reaches the `db.Query` the last connection might be occupied by the transaction and therefore blocks until a connection is available, which might never happen.

So how do you fix this? First of all, all queries should be run either outside or inside a transaction for a specific part of your code. Even if you run a transaction block and a non-transaction block concurrently, the non-transaction block will not be blocked by the transaction forever (but the non-transaction block might need to wait for the other part to finish). Additionally, you can use a linter or another tool to make sure all queries are run completely inside or outside of transactions.

I usually write integration tests against the database. If you do the same, you can configure the connection pool size to make sure the tests will only use a single connection. That way the tests will run into a deadlock should you have forgotten to use a transaction somewhere. You can easily configure that inside the `TestMain` function for a package.

```
func TestMain(m *testing.M) {
	db.SetMaxOpenConns(1) // db is the *sql.DB created somewhere
}
```

I hope this helps you to prevent some nasty deadlock bugs. I found quite a few in a larger code base by limiting the connection pool. In production, you should use multiple connections to speed up things of course.

* * *

Would you like to see more? Read my blog articles on [Emvi](https://emvi.com/blog?ref=marvinblum.de), my project page on [GitHub](https://github.com/Kugelschieber) or send me a [mail](mailto:marvin@marvinblum.de).

This page uses [concrete](https://concrete.style/) for styling. Check it out!

This page does not use cookies. [Legal](https://marvinblum.de/legal)
