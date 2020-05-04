write-output("packer_build_name is automatically set for you, or you can set it in your builder variables; the default for this builder is: " + $env:packer_build_name )
write-output("remember that escaping variables in powershell requires backticks; for example var1 from our config is " + $env:var1 )
write-output("likewise, var2 is " + $env:var2 )
write-output("and var3 is " + $env:var3 )

