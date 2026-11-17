# How does it work

Server have a core part, application instance and adapters.

Core is the main part of the application - the foundation. It's how modules communicate with application instance -
through the core. Core contain several buses - command, query and event bus. Command change the data, query bus - only
ask for data, event bus - need to publish independent events that other modules might be subscribed to. Why separate
command and query bus - because of database and transactions

Application instance. This thing assemble everything togather - modules, outer world things ( db, cache, email sender,
notification sender ) and a core itself. Why separate core and application instance? Testability and different
responsibilities. Core is a transport ( maybe this name will fitt better ) while Instance is what link everything togather.
So when we provide an App to the cli command handler - it's safe because cli command handler should not know about everything
inside the application. It should know only about transport interfaces and how to call the command.

Adapters - are CLI, rest api, graphql api, static web pages, rpc or grpc and so on. Each adapter know ONLY about public
modules contract - how to call the command. But it should not know about internal implementation.