# Dependencies

Building and testing Gateway requires:

 - A local clone of `git@github.com:AnyPresence/gateway.git`
 - [Go](#go)
 - [NPM](#npm)
 - [ember and ember-cli](#ember)
 - [rubygems](#rubygems) and bundler
 - [sqlite](#sqlite)
 - [postgresql](#postgresql)

Mac OS users will want to use [`homebrew`](http://brew.sh) to install packages.

## Make sure you are not logged in as `root` when installing packages via `npm` or `gem`!

---

## Go

### Installing

 > Mac OS

```bash
brew install go
```

 > Ubuntu Linux

```bash
apt-get install golang
```

### Setup

```bash
mkdir ~/go
echo 'export GOPATH=${HOME}/go' >> ~/.bashrc
echo 'export PATH=${PATH}:${GOPATH}/bin' >> ~/.bashrc
source ~/.bashrc
```

## NPM

 > Mac OS

```bash
brew install node # Should 'just work' without further finagling.
```

 > Ubuntu Linux

```bash
curl -sL https://deb.nodesource.com/setup_0.12 | sudo bash -
sudo apt-get install -y nodejs
sudo ln -sf `which nodejs` /usr/bin/node
mkdir ~/.npm_packages
npm config set prefix ~/.npm_packages
```

## Ember

```bash
npm install -g ember
npm install -g ember-cli
```

## Rubygems

### Installing

 > Mac OS

```bash
brew install chruby
brew install ruby-install
```

 > Ubuntu Linux

Follow the install instructions for [chruby](https://github.com/postmodern/chruby#install) and [ruby-install](https://github.com/postmodern/ruby-install#install).

### Configuring

```bash
source /usr/local/share/chruby/chruby.sh
chruby ruby-2.2.2
echo 'source /usr/local/share/chruby/chruby.sh' >> ~/.bashrc
echo 'chruby ruby-2.2.2' >> ~/.bashrc
```

### Get Bundler

```bash
gem install bundler
```

## SQLite

 > Mac OS

```bash
brew install sqlite
```

 > Ubuntu Linux

```bash
apt-get install -y sqlite
```

## PostgreSQL

### Installing

 > Mac OS

```bash
brew install postgresql
```

 > Ubuntu Linux

```bash
apt-get install -y postgresql
```

### Configuring

 > Mac OS X

```bash
echo 'pg_ctl -D /usr/local/var/postgres -l /usr/local/var/postgres/server.log start' > /usr/local/bin/pg_start
echo 'pg_ctl -D /usr/local/var/postgres stop -s -m fast' > /usr/local/bin/pg_stop
chmod +x /usr/local/bin/pg_start
chmod +x /usr/local/bin/pg_stop
```

Mac users can now use `pg_start` and `pg_stop` to start and stop their user's Postgres instance.

 > Ubuntu Linux

```bash
sudo service postgresql start
sudo su postgres
createuser --superuser <your username>
psql -x -c "ALTER ROLE <username> WITH PASSWORD NULL;"
exit
```

FIXME: This is suboptimal.  It would be better to create a userlevel service instead of a global user with superuser and no password.
