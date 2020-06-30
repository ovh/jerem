# Jerem

Jerem is a [golang](https://golang.org/doc) bot that scrap [JIRA](https://www.atlassian.com/software/jira) project to extract Metrics. Those Metrics can then be send to a Warp 10 Backend.

## Jerem and Metrics

Jerem was build to send data to the [OVH Metrics Data Platform](https://www.ovh.com/fr/data-platforms/metrics/).
Copy the `config.sample.yml` file into a local or edit directly your `config.yml`.
Jerem only require the following valid configuration keys:

```yaml
metrics:
  url: https://warp.gra1.metrics.ovh.net
  token: METRICS_READ_TOKEN  # On the Metrics Data Platform, you can follow this [documentation](https://docs.ovh.com/gb/en/metrics/order/) to get a valid token.
```

## Scrap a JIRA Project

Once your datasource is correctly set, you will want to add JIRA project to JEREM to scrap.
Copy the `config.sample.yml` file into a local or edit directly your `config.yml`.
First set your will need to configure JIRA based on the following configure keys:

```yaml
jira:
  username: jira.bot
  password: 123456789
  url: https://jira.com
```

If you have custom closed JIRA status for a ticket, you can set the following jira config key:

```yaml
jira:  
  closed.statuses:
    - Done
```

By default the closed JIRA status are `(Resolved, Closed, Done)`. A list is expected for this optional parameter.

Then adding the project is simply done by editing the `projects` key. A single JIRA project require only two keys in the configuration file: it's project name and it's board id.

```yaml
projects:
  - name: OB  # The JIRA board project name
    board: 0  # The JIRA board id (can be found in the board URL)
    jql_filter: component = test # Optional parameter to filter JIRAs inside a project
    label: OB_test # Optional parameter to override the project name with a custom label
```

Then you can simply add a second project:

```yaml
projects:
  - name: OB
    board: 0
    jql_filter: component = test
    label: OB_test
  - name: K8S
    board: 1  
```

## Compile and run jerem

You will need to have Golang set-up locally: check their [golang installation step](https://golang.org/doc/install).
Once jerem is well configure, you can simply run:

```sh
# Compile
make compile

# Run (will load local config.yaml file)
./bin/jerem

# Run with a remote config file
./bin/jerem --config /Path/to/config.yaml
```

## Contributing

Instructions on how to contribute to Jerem are available on the [Contributing](./CONTRIBUTING.md) page.

## License

Jerem is released under a [3-BSD clause license](./LICENSE).

## Grafana's dashboards

In the grafana folder we are going to integrate all dashboards computed from Jerem metrics:

- program_management.json : the template for the grafana project management overview.

This dashboard needs a valid Warp 10 data source set up, with an `RTOKEN` variable set.
To set up custom project for the program management dashboard, you need to update the `Project` variable. The value is a WarpScript string list. A valid value can be `[ 'SAN' 'OTHER' ]`. This render a dashboard in which we can switch on the `SAN` or to `OTHER` program management dashboard.

## About Jerem

Here we keep track of all blog posts about the Agile methodology related to Jerem:

- [The birth of agile telemetry at OVHcloud – Part I](https://www.ovh.com/blog/the-birth-of-agile-telemetry-at-ovhcloud-part-i/)
- [Jerem: An Agile Bot](https://www.ovh.com/blog/jerem-an-agile-bot/)
- [Agile telemetry at OVHCloud – Part II](https://www.ovh.com/blog/agile-telemetry-at-ovhcloud-part-ii/)
- [Agile telemetry at OVHCloud – Part III](https://www.ovh.com/blog/agile-telemetry-at-ovhcloud-part-iii/)

## Get in touch

- Gitter: [metrics](https://gitter.im/ovh/metrics)
