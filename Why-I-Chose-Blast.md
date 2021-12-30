# Why I chose Blast!

This is a review of the different search engines available on GitHub. 

## Introduction

I am building multiple subject-specific rss search engines. I aggregate rss feeds on a particular topic and allow the user to search them.  For that I need a search engine.  

I want one that is open source, has the right functionality, good performance, scalable, and easy to modify.  
There are so many on the market.  So I organized my search for a search engine by implementation language. 

## Python
I am primarily a Python developer, so first I looked at Python-bsed solutions.  
Quickly I found Whoosh.  It has two great features.  First I can specify the schema.  (Blast also lets me define a schema, but many other search engines just assume a single blob of text.)  Whoosh even lets me weigh matches on certain fields, like the, title higher.  And I can easily customize the ranking algorithm. 

But I decided against Whoosh.  Not at all clear how to scale it.  And they have this huge problem that the original developer is in Mercurial, but the community edition is on Github, and they have not managed to sync up the versions with PyPi.  So Whoosh is out. 

## Ruby
I am increasingly moving to Ruby. On github I quickly found Saus Engine. 
What a brilliant introduction to all of the search engine issues. 
https://github.com/sausheong/saushengine.v1

But I would have to also install MySQL, so I quickly moved on. 

## Javascript
There is this beautiful README for  a mature-looking Javascript search engine, with apis in every language. 
https://github.com/meilisearch/MeiliSearch
 
  I particularly liked that they reviewed the competition.  https://docs.meilisearch.com/learn/what_is_meilisearch/comparison_to_alternatives.html#a-quick-look-at-the-search-engine-landscape.  
It means that they are self confident. But they did not mention Blast, maybe because they are scared of us. 


But I could not find a way to define a schema, I don’t like Javascript (I am a Python developer), and it was not clear how to scale.  So once again, I moved on. 

## Rust
The reason to use rust is for performance.  There is this excellent article about why Discord moved from GoLang to Rust.  https://blog.discord.com/why-discord-is-switching-from-go-to-rust-a190bbca2b1f?gi=721b3f9044ba

(Note the tracking cookie in the URL)
That makes sense for very high volume sites.  But Rust is hugely difficult conceptually, which I think gets in the way of adding new functionality.  IMHO GoLang has a much better model for concurrency, and a better track record for wonderful applications. 

## GoLang

I am increasingly fond of applications written in GoLang.  I have had a great experience with the Caddy Web Server.  Way easier to configure than NGINX, and it offers so much more functionality, there is really no comparison.  

Of course handling errors manually in GoLang is painful, compared to Python or Ruby,  but I do not expect to be developing in GoLang, just consuming the free applications. 

Just like Caddy, Blast looks like it has wonderful simple functionality.   
One of my problems is that I did not even know what functionality a search engine should have.  
I think Blast  got it just right.   Now I use the Blast interface as what I expect from a search engine.  To better understand the blast functionality, please watch this Bleve video. 
https://www.youtube.com/watch?v=DfbRTXE5n4Y

You can read more about Bleve explorer here.  https://github.com/blevesearch/bleve-explorer

https://en.wikipedia.org/wiki/Tf%E2%80%93idf

Is Blast perfect?  Nothing is.  Not clear how to modify the ranking algorithms.   Although I am sure it is possible.  And we need to add a link to a recommended Javacript 
client library with term suggestins.  But overall Blast looks really excellent. It has an excellent interface.  It scales.  It is free and open source.   

This was my very rushed review of the market.  I am sure I made some mistakes.  Please edit this document and add what you have learned. 



 
