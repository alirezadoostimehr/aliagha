# Viper 

The Viper package is used in the project to handle configuration settings. It allows us to read configuration values from various sources such as files, environment variables, and command-line flags. This documentation provides an overview of how to initialize and use the Viper package in our project.


## Initialization
1. Open your terminal or command prompt.
2. Run the following command to install Viper: `go get -u github.com/spf13/viper`.
3. Wait for the installation process to complete. It will download and install the necessary files.
4. Now we need to provide the configuration file path, name, and type. The Init function handles the initialization process. Here's an example of how to use it:

```type Params struct {
	FilePath string
	FileName string
	FileType string
}

func Init(param Params) (*Config, error) {
	viper.SetConfigType(param.FileType)
	viper.AddConfigPath(param.FilePath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}

	// Configuration structure initialization...

	return &Config{
		// Configuration field assignments...
	}, nil
}
```
In the Params struct, we provide the FilePath, FileName, and FileType values to specify the location and type of your configuration file. The Init function sets up the Viper package by specifying the configuration type, adding the configuration path, and reading the configuration file using `viper.ReadInConfig()`.

## Usage
Once the Viper package is initialized, we can access the configuration values using `viper.GetString(key)` or `viper.GetInt(key)` methods, where key is the configuration key we want to retrieve. Here's an example of accessing configuration values:
```func Init(param Params) (*Config, error) {
	// Viper initialization code...

	redis := &Redis{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
	}

	// Other configuration value retrievals...

	return &Config{
		Redis: *redis,
		// Other configuration assignments...
	}, nil
}
```
In the example above, the redis.host, redis.port, and redis.password values are retrieved using `viper.GetString()` and `viper.GetInt()` methods and assigned to the Redis struct fields.

We can access other configuration values in a similar manner by specifying the respective keys.

## Features
#### The Viper package provides several features to enhance the configuration handling in our project:

* Configuration Sources: Viper supports reading configuration values from various sources, including files (JSON, YAML, TOML, etc.), environment variables, and command-line flags.
* Automatic Environment Variable Binding: Viper can automatically bind configuration values to environment variables. By following a specific naming convention, we can override configuration values using environment variables.
* Default Values: Viper allows you to set default values for configuration keys. If a configuration value is not found, Viper will fall back to the default value.
* Watch and Hot-Reload: Viper provides the ability to watch configuration files for changes and automatically reload the configuration when modifications occur. This is useful for dynamically updating configuration settings during runtime.
* Nested Configurations: Viper supports nested configurations, allowing you to organize our configuration values into hierarchical structures.



