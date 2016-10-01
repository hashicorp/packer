winrmcp
=======

Copy files to a remote host using WinRM

[![wercker status](https://app.wercker.com/status/00e746084a49de5d91654b23cdf97a5e/m "wercker status")](https://app.wercker.com/project/bykey/00e746084a49de5d91654b23cdf97a5e)

Example:

    make
    bin/winrmcp -help
    bin/winrmcp -user=vagrant -pass=vagrant ~/Downloads/fortune.jpg C:/Cookies/fortune.jpg
