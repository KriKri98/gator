Gator Blogaggregator:
This project has been created in line with the course on boot.dev.

How to use:
You need GO an Postgres installed in Linux.
To install the gator program, you can use go install https://github.com/KriKri98/gator
With postgres installed, you need to create a new database, e.g. createdb gator
Run the migration files in sql/schema/ against that database with somethin like goose.
After installation, you need to set up a config file named ".gatorconfig.json" in your root directory.
You can find your root with ~/
The config file contents need to be:
Here you need to set the connection string to your database.
{"db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable","current_user_name":""}

You may need to change the db_url to connect correctly to the database.

After correct installation, you can use the following commands, all starting with gator:
register: needs a username to register, registers a new user and set that user as active
login: needs a username, switches the user to the new one
reset: clears all data
follow: needs a url of a RSS-feed, mark the feed as followed by the current user
unfollow: needs a url of a RSS-feed, marks the feed as unfollowed by the current user
agg: needs a time duration like 10s, connects to the feeds in the specified intervall and gets the information from the feeds
browse: has an optional limit, prints the feeds up to the specified limit








