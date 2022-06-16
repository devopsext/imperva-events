# Imperva events exporter

Simple application which poll events from Imperva and send them to different outputs. Currently, supported outputs are: 

- Stdout
- Slack
- Grafana annotations

## Usage

```
Usage:
  imperva-events [flags]

Flags:
      --account-id string        Imperva Account ID
      --api-id string            Imperva API ID
      --api-token string         Imperva API Token
      --debug                    Enable debug logging
      --grafana-api-key string   Grafana API Key
      --grafana-url string       Grafana URL
  -h, --help                     help for imperva-events
      --init-interval int        Imperva Init interval (minutes) (default 600)
      --poll-interval int        Imperva Poll interval (seconds) (default 10)
      --slack-channel string     Slack channel
      --slack-token string       Slack token
```

### Built With

* [Golang](https://go.dev/)

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE.txt` for more information.
