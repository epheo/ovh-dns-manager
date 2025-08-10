package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"ovh-dns-manager/internal/config"
	"ovh-dns-manager/internal/ovh"
	"ovh-dns-manager/internal/sync"
)

var (
	credentialsFile string
	configFile      string
	domain          string
	outputFile      string
	dryRun          bool
)

// setupOVHClient loads credentials and creates OVH client
func setupOVHClient(credentialsFile string) (*ovh.Client, error) {
	creds, err := config.LoadOVHCredentials(credentialsFile)
	if err != nil {
		return nil, err
	}

	return ovh.NewClient(creds)
}

// resolveValueWithEnvFallback resolves a flag value with environment variable fallback
func resolveValueWithEnvFallback(flagValue, envValue, flagName, envVarName string) (string, error) {
	if flagValue == "" {
		if envValue != "" {
			return envValue, nil
		}
		return "", fmt.Errorf("%s is required (use --%s flag or set %s environment variable)", flagName, flagName, envVarName)
	}
	return flagValue, nil
}

var rootCmd = &cobra.Command{
	Use:   "ovh-dns-manager",
	Short: "Manage OVH DNS zones via YAML configuration",
	Long:  "A tool to export and apply DNS zone configurations to OVH using YAML files",
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export DNS zone configuration to YAML file",
	Long:  "Fetch DNS records from OVH API and generate a YAML configuration file",
	RunE:  runExport,
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply DNS zone configuration from YAML file",
	Long:  "Read YAML configuration and sync DNS records to OVH (one-way sync)",
	RunE:  runApply,
}

func init() {
	credentialsPath, envDomain, configPath := config.LoadAppConfig()
	
	rootCmd.PersistentFlags().StringVarP(&credentialsFile, "credentials", "c", credentialsPath, "OVH credentials file")
	
	exportCmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain to export (required)")
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output YAML file (default: {domain}.yaml)")
	
	// Make domain flag not required if OVH_DOMAIN env var is set
	if envDomain == "" {
		exportCmd.MarkFlagRequired("domain")
	}

	applyCmd.Flags().StringVarP(&configFile, "config", "f", "", "DNS configuration YAML file (required)")
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show changes without applying them")
	
	// Make config flag not required if OVH_CONFIG_PATH env var is set
	if configPath == "" {
		applyCmd.MarkFlagRequired("config")
	}

	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(applyCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	_, envDomain, _ := config.LoadAppConfig()
	
	var err error
	domain, err = resolveValueWithEnvFallback(domain, envDomain, "domain", "OVH_DOMAIN")
	if err != nil {
		return err
	}

	client, err := setupOVHClient(credentialsFile)
	if err != nil {
		return err
	}

	syncer := sync.NewSyncer(client, false)
	zone, err := syncer.ExportZone(domain)
	if err != nil {
		return err
	}

	if outputFile == "" {
		outputFile = domain + ".yaml"
	}

	if err := config.SaveDNSZone(zone, outputFile); err != nil {
		return err
	}

	log.Printf("Exported %d DNS records for domain %s to %s", len(zone.Records), domain, outputFile)
	return nil
}

func runApply(cmd *cobra.Command, args []string) error {
	_, _, envConfigPath := config.LoadAppConfig()
	
	var err error
	configFile, err = resolveValueWithEnvFallback(configFile, envConfigPath, "config", "OVH_CONFIG_PATH")
	if err != nil {
		return err
	}

	zone, err := config.LoadDNSZone(configFile)
	if err != nil {
		return err
	}

	client, err := setupOVHClient(credentialsFile)
	if err != nil {
		return err
	}

	syncer := sync.NewSyncer(client, dryRun)
	result, err := syncer.SyncZone(zone)
	if err != nil {
		return err
	}

	result.PrintSummary()

	if result.HasErrors() {
		return fmt.Errorf("sync completed with %d errors", len(result.Errors))
	}

	if dryRun && result.HasChanges() {
		log.Println("Dry run completed. Use --dry-run=false to apply changes.")
	} else if result.HasChanges() {
		log.Println("DNS zone sync completed successfully")
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}