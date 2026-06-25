test {
  parallel = true
}

run "provider_example_configures" {
  command = plan

  module {
    source = "./examples/provider"
  }
}
