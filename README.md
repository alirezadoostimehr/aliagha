# aliagha
Biggest competitor of alibaba.ir! (Quera bootcamp final project)
* Introduction:
 The purpose of this project is to develop a selling ticket website that allows users to browse and purchase tickets for various events. The website integrates with a mock API to retrieve event data and provides a seamless ticket buying experience. The project utilizes the Echo framework, Viper for configuration management, Validator for input validation, JWT Go for authentication, and Migrate for database migrations.
 * Version control
	* [branch and merging](https://github.com/alirezadoostimehr/aliagha/blob/init-document/versioncontrol/branchandmerg.md)
	
 * Packages
 	* [cobra](https://github.com/alirezadoostimehr/aliagha/blob/init-document/packages/cobra.md)
 	* [viper](https://github.com/alirezadoostimehr/aliagha/blob/init-document/packages/viper.md)
 	* [echo](https://github.com/alirezadoostimehr/aliagha/blob/init-document/packages/echo.md)
 	* [validator v10](https://github.com/alirezadoostimehr/aliagha/blob/init-document/packages/validator.md)
 	* [migrate](https://github.com/alirezadoostimehr/aliagha/blob/init-document/packages/migrate.md)
 	* [jwt-go](https://github.com/alirezadoostimehr/aliagha/blob/init-document/packages/jwtgo.md)
 * Database
 	* Database design process followed by our team for the project. It provides insights into how we understood the entities, tables, and structs in order to create an efficient and well-structured database.
 	* Understanding Entities
To begin the database design process, we thoroughly analyzed the project requirements, user stories, and system diagrams. This allowed us to identify the key entities involved in the system. We engaged in discussions with stakeholders to gain a deep understanding of the project's domain and the relationships between entities.

	* Mapping Entities to Tables
Once we had a clear understanding of the entities, we proceeded to map them to database tables. For each entity, we identified the relevant attributes and determined their data types. We also considered any constraints or validations that needed to be applied to the data.
To establish relationships between tables, we analyzed the associations between entities. We identified one-to-one, one-to-many, and many-to-many relationships, and implemented them using primary and foreign keys. This ensured data integrity and maintained proper referential integrity between tables.

	* Creating Structs
Based on the identified entities and their attributes, we proceeded to create corresponding structs in our codebase. These structs served as representations of the database tables and facilitated seamless interaction with the database through object-relational mapping (ORM) frameworks.
We carefully matched the attributes of the structs with the columns of the corresponding tables, ensuring consistency and accurate data retrieval and manipulation.
