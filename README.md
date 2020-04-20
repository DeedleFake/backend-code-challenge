bcc
===

bcc is a demo social media API backend.

API Documentation
-----------------

cmd/bcc exposes a REST API server. The server's GET endpoints take simple query parameters, while POST endpoints expect a JSON body in the request.

#### Endpoints

* `/timeline`:
  * `GET`:
		* Returns the "timeline" for a user.
		* Params:
			* `user_id`: (number, required), the ID of the user whose timeline is being fetched
			* `start`: (number, default 0), the starting element of the timeline for pagination purposes
			* `limit`: (number, default 10, max 100), the number of elements to return
* `/post`:
  * `GET`:
		* Fetches a post and its comments.
		* Params:
			* `post_id`: (number, required), the ID of the post being fetched
	* `POST`:
	  * Creates a new post.
		* Params:
		  * `user_id`: (number, required), the ID of the user making the post
			* `title`: (string, required), the title of the post
			* `body`: (string), the body of the post
