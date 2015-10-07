# Dependencies

Building and testing Gateway requires:

 - A local clone of `git@github.com:AnyPresence/gateway.git`
 - [Go](#go)
 - [NPM](#npm)
 - [ember and ember-cli](#ember)
 - [rubygems](#rubygems) and bundler
 - [sqlite](#sqlite)
 - [postgresql](#postgresql)
 - [java](#java)

Note that gateway uses git submodules.

```bash
git clone git@github.com:AnyPresence/gateway.git
cd gateway
git submodule init
git submodule update
```

Mac OS users will want to use [`homebrew`](http://brew.sh) to install packages.

## Make sure you are not logged in as `root` when installing packages via `npm` or `gem`!

---

## Go

### Installing

 > Mac OS

```bash
brew install go
```

 > Linux

```bash
export GOVERSION="1.5.1.linux-amd64" # Or whatever platform and version
wget "https://storage.googleapis.com/golang/go${GOVERSION}.tar.gz"
tar -C /usr/local -xzf "go${GOVERSION}.tar.gz"
echo 'PATH="${PATH}:/usr/local/go/bin"' >> ~/.bashrc
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


## Java

You will need JDK 1.8 or greater in order to be able to connect to SOAP remote endpoints.

### Installing

 > Mac OS

```bash
brew update
brew tap caskroom/cask
brew install Caskroom/cask/java
```

 > Ubuntu Linux
 > (You may need to log out and back in after installing.)

```bash
add-apt-repository ppa:webupd8team/java
apt-get update
apt-get install -y oracle-java8-installer oracle-java8-set-default
```

### Verifying

Try running `java -version`.  You should see something like the following.  Make sure the
first line matches the expected version number:  *java version "1.8.0_XX"*

```bash
$ java -version
java version "1.8.0_45"
Java(TM) SE Runtime Environment (build 1.8.0_45-b14)
Java HotSpot(TM) 64-Bit Server VM (build 25.45-b02, mixed mode)
```

If the above is successful, you should have the correct version installed.  All that is left
is to make sure that wsimport can be found correctly on your path.  Try running `wsimport -version`.
You should see something similar to the following.  The exact version number for wsimport is not
important -- just make sure that wsimport is available on the path, and that the instructions for
installing Java with the correct version are completed successfully.

```bash
$ wsimport -version
wsimport version "2.2.9"
```
