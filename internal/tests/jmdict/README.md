# jmdict

This directory tests generating a Go struct for the [JMdict
project](https://www.edrdg.org/jmdict/j_jmdict.html).

Since the JMdict project's data file is extremely large and includes profanity,
it is not included in the repo by default.

To enable this test you must first download the data file by running this
command in the the project's root directory:

```console
$ curl -o internal/tests/jmdict/testdata/JMdict.gz ftp://ftp.edrdg.org/pub/Nihongo/JMdict.gz
```