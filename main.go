package main

import (
	"fmt"
	"log"
	"os"

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
	rootCmd.PersistentFlags().StringVarP(&credentialsFile, "credentials", "c", "ovh-credentials.yaml", "OVH credentials file")
	
	exportCmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain to export (required)")
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output YAML file (default: {domain}.yaml)")
	exportCmd.MarkFlagRequired("domain")

	applyCmd.Flags().StringVarP(&configFile, "config", "f", "", "DNS configuration YAML file (required)")
	applyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show changes without applying them")
	applyCmd.MarkFlagRequired("config")

	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(applyCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	creds, err := config.LoadOVHCredentials(credentialsFile)
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	client, err := ovh.NewClient(creds)
	if err != nil {
		return fmt.Errorf("failed to create OVH client: %w", err)
	}

	syncer := sync.NewSyncer(client, false)
	zone, err := syncer.ExportZone(domain)
	if err != nil {
		return fmt.Errorf("failed to export zone: %w", err)
	}

	if outputFile == "" {
		outputFile = domain + ".yaml"
	}

	if err := config.SaveDNSZone(zone, outputFile); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	log.Printf("Exported %d DNS records for domain %s to %s", len(zone.Records), domain, outputFile)
	return nil
}

func runApply(cmd *cobra.Command, args []string) error {
	zone, err := config.LoadDNSZone(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	for i := range zone.Records {
		if err := config.ValidateDNSRecord(&zone.Records[i]); err != nil {
			return fmt.Errorf("invalid DNS record %d: %w", i, err)
		}
	}

	creds, err := config.LoadOVHCredentials(credentialsFile)
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	client, err := ovh.NewClient(creds)
	if err != nil {
		return fmt.Errorf("failed to create OVH client: %w", err)
	}

	syncer := sync.NewSyncer(client, dryRun)
	result, err := syncer.SyncZone(zone)
	if err != nil {
		return fmt.Errorf("failed to sync zone: %w", err)
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
		os.Exit(1)
	}
}