#!/home/ben/software/install/bin/perl
use warnings;
use strict;
use utf8;
use FindBin '$Bin';
use Perl::Build;
perl_build (
    makefile => 'Makefile',
);
