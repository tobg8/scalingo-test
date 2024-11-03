# Canvas for Backend Technical Test at Scalingo (THEO BELOEIL GUIA)

## Before running the project

- Create a `.env` file with the `GITHUB_TOKEN` environment variable
 This is needed because we crush github api rate limit quickly using go multithreading.

 To get a token, go to <https://github.com/settings/tokens> and generate a token with the `public_repo` scope.

You good to go running the project with `docker compose up`.

## Project requirements

- [x] Use Go
- [x] Use Docker
- [x] Use concurrency
- [x] Fetch repositories from the github API based on language and multiple filters
- [x] Fetch requested language data of each repository
- [x] Return the aggregated data in the response

- [] implement the filter order
- [] impelment the filter sort

## Points to consider

- The request MUST contain a `language` query parameter, this is the point to me, do a search public repositories on a specific language.
- The response IS paginated, 100 repositories per page (but can be more or less, *per_page* parameter)

## Usage of API

The API uses the same filters as the one used in github API.
<https://docs.github.com/fr/search-github/searching-on-github/searching-for-repositories>

a simple query looks like this:

<http://localhost:5000/repos?q=language:javascript:size:1..10>

- a filter is a key:value pair, the key is the filter name and the value is the filter value (can be a range, a number, a date, etc.).
- the `+` character is used to separate filters.

I did not implement all filters, but most of them are supported, the rest can be implemented easily [Filters available](#filter-support).

⚠️ I did not implement a check of  of `license` and `language` filters validity, so if you provide a wrong license of language filter, it will return a poor error message, I should the fetch the data from github to check if the license or language is valid and make a proper error message. (or create a map of valid licenses and languages) but this is hacky and won't allow us to scale accordingly with the api ⚠️

The APi offers only one endpoint:

- `GET /repos`

## Filter Support

The query parameters are:

<http://localhost:5000/repos?q=key:value+key:value&per_page=100&page=1>

- *q* - the query string

___

- *size*     - 1..10||>=10||<=10||:20
- *topics* - 1..10||>=10||<=10||:20
- *stars*    - 1..10||>=10||<=10||:20
- *followers* - 1..10||>=10||<=10||:20
- *forks*     - 1..10||>=10||<=10||:20

- *license* - MIT||GPL||BSD
- *language* - javascript || python || go || rust

- *created* - >=2024-01-01||<=2024-01-01||:2024-01-01
- *pushed* - >=2024-01-01||<=2024-01-01||:2024-01-01

___

- *per_page* - number of items per page (default: 100, max 100)

___

- *page* - number of the page (default: 1)

## Examples

- search public repositories with the word `scalingo` (in name, description or topics) and the language `javascript` and the size of the repository is between 1 and 10 KB
  - <http://localhost:5000/repos?q=scalingo+language:javascript+size:1..10>
- search public repositories with the language `javascript` and the size of the repository is smaller or equal to 100 KB
  - <http://localhost:5000/repos?q=language:javascript+size:<=100>
- search public repositories with the language `go` and the followers is 10 and stars is greater than 10
  - <http://localhost:5000/repos?q=language:go+followers:10+stars:>10>
- search public repositories with the language `rust` and the size of the repository is between 1 and 10 KB and the number of stars is greater or equal to 10 and the number of followers is greater or equal to 100
  - <http://localhost:5000/repos?q=language:rust+size:1..10+stars:>=10+followers:>=100>

## Project structure

I do use the clean architecture pattern, so the project is divided into 4 layers in the `src` folder:

- `controllers` contains the handlers for the API endpoints.
- `repositories` contains the implementation of the repository interface.
- `usecases` contains the business logic of the API.
- `models` contains the data models of the API.

You can find the entry point of the API in `main.go`.

Clean architecture is a pattern that helps to separate the concerns of the application, it helps to make the code more testable and more maintainable, allowing to write unit tests easily and perform mocking easily.
This pattern is usefull for separation of concerns, it helps to make the code more modular and easier to maintain.
