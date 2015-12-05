#!/home/ben/software/install/bin/perl
use warnings;
use strict;
use utf8;
use FindBin '$Bin';
use HTML::Valid::Tagset ':all';
for my $tag (@allTags) {
    if ($isHTML5{$tag}) {
	print $tag;
	if ($isBlock{$tag}) {
	    print " block";
	}
	if ($isInline{$tag}) {
	    print " inline";
	}
	if ($isHeadElement{$tag} && ! $isBodyElement{$tag}) {
	    print " head-only";
	}
	if ($isTableElement{$tag}) {
	    print " table";
	}
	if ($isObsolete{$tag}) {
	    print " OBSOLETE!\n";
	}
	if ($isFormElement{$tag}) {
	    print " form";
	}
	if ($emptyElement{$tag}) {
	    print " empty";
	}
	if ($optionalEndTag{$tag}) {
	    print " end-optional";
	}
	print "\n";
    }
}
