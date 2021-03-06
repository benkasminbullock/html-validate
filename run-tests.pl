#!/home/ben/software/install/bin/perl
use warnings;
use strict;
use utf8;
use FindBin '$Bin';
use Deploy 'do_system';
use File::Slurper 'read_text';
use Test::More;
# Switch on or off debugging flags
my $verbose = 1;
# globals
my $outtmpl = "$Bin/validate-test-out";
my $output = "$outtmpl.$$";
my $errors = "$outtmpl.errors.$$";
clean ();
checkold ();
my @files = ("t/mr-old.html");
my $bin = "$Bin/validate";
if (! -f $bin || ! -x $bin) {
    if ($verbose) {
	print "Rebuilding $bin.\n";
    }
    do_system ("make validate", $verbose);
}
for my $file (@files) {
    die unless -f $file;
    do_system ("$bin $file > $output 2> $errors");
    ok (-f $output, "got output");
    ok (! -s $errors, "no errors");
    my $text = read_text ($output);
    my @lines = split /\n/, $text;
    my $n = 0;
    for (@lines) {
	$n++;
	like ($_, qr!$file:[0-9]+: !, "line $n is OK");
    }
    clean ();
}
done_testing ();
exit;
sub clean
{
    for my $file ($output, $errors) {
	if (-f $file) {
	    unlink $file or die $!;
	}
    }
}
sub checkold
{
    my @outfiles = <$outtmpl.*>;
    if (@outfiles) {
	warn "Old output files @outfiles remain";
    }
}
