package cmd

import (
	"fmt"
	"strings"
	"time"

	"devsentry/internal/clean"
	sizeutil "devsentry/internal/size"

	"github.com/spf13/cobra"
)

var cleanApply bool
var cleanTimeout time.Duration
var cleanImages bool
var cleanContainers bool
var cleanCache bool
var cleanAll bool
var cleanJSON bool

type cleanOutput struct {
	Mode    string     `json:"mode"`
	Plan    clean.Plan `json:"plan"`
	Applied bool       `json:"applied"`
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up Docker resources",
	Args:  cobra.NoArgs,
	RunE:  runCleanCmd,
}

func runCleanCmd(cmd *cobra.Command, args []string) error {
	if err := validateCleanFlags(cleanApply, cleanAll, cleanImages, cleanContainers, cleanCache); err != nil {
		return err
	}

	hasScope := cleanImages || cleanContainers || cleanCache
	selectAll := cleanAll || !hasScope
	opts := clean.Options{
		Timeout:    cleanTimeout,
		Images:     selectAll || cleanImages,
		Containers: selectAll || cleanContainers,
		BuildCache: selectAll || cleanCache,
		All:        cleanAll,
	}
	plan, err := clean.BuildPlan(cmd.Context(), opts)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	mode := "DRY-RUN"
	if cleanApply {
		mode = "APPLY"
	}

	if cleanJSON {
		if cleanApply {
			if err := clean.Apply(cmd.Context(), opts); err != nil {
				return err
			}
		}
		return writeJSON(out, cleanOutput{Mode: mode, Plan: plan, Applied: cleanApply})
	}

	style := newOutputStyle(out)
	modeLabel := style.warning(mode)
	if cleanApply {
		modeLabel = style.success(mode)
	}
	fmt.Fprintf(out, "%s\n", style.title(fmt.Sprintf("Docker cleanup (%s)", modeLabel)))

	if opts.Images {
		fmt.Fprintf(out, "  Dangling images:    %d (%s)\n", len(plan.DanglingImages), formatBytes(sumImageBytes(plan.DanglingImages)))
		if opts.All {
			fmt.Fprintf(out, "  Other unused images: %d (%s)\n", len(plan.UnusedImages), formatBytes(sumImageBytes(plan.UnusedImages)))
			for _, image := range plan.UnusedImages {
				fmt.Fprintf(out, "    %s  %s  %s\n", image.Name, image.ID, image.SizeLabel)
			}
		}
	}
	if opts.Containers {
		fmt.Fprintf(out, "  Stopped containers: %d (%s)\n", len(plan.StoppedContainers), formatBytes(sumContainerBytes(plan.StoppedContainers)))
	}
	if opts.BuildCache {
		if plan.BuildCache.Available {
			suffix := ""
			if plan.BuildCache.Approx {
				suffix = " estimated"
			}
			fmt.Fprintf(out, "  Build cache:        %s%s\n", plan.BuildCache.SizeLabel, suffix)
		} else {
			fmt.Fprintln(out, "  Build cache:        unknown")
		}
	}
	fmt.Fprintln(out)

	if plan.EstimateApprox && plan.EstimatedBytes == 0 {
		fmt.Fprintln(out, "Estimated reclaim: unknown")
	} else {
		prefix := ""
		if plan.EstimateApprox {
			prefix = "~"
		}
		reclaim := fmt.Sprintf("%s%s", prefix, formatBytes(plan.EstimatedBytes))
		fmt.Fprintf(out, "Estimated reclaim: %s\n", style.success(reclaim))
	}

	if !cleanApply {
		fmt.Fprintln(out, style.warning("Dry-run only. Nothing was removed."))
		fmt.Fprintf(out, "Apply with: devsentry clean %s --apply\n", cleanScopeArgs(cleanAll, cleanImages, cleanContainers, cleanCache))
		return nil
	}

	if err := clean.Apply(cmd.Context(), opts); err != nil {
		return err
	}

	fmt.Fprintln(out, style.success("Cleanup completed."))
	return nil
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanApply, "apply", false, "Execute the cleanup commands")
	cleanCmd.Flags().BoolVar(&cleanImages, "images", false, "Clean dangling images")
	cleanCmd.Flags().BoolVar(&cleanContainers, "containers", false, "Clean stopped containers")
	cleanCmd.Flags().BoolVar(&cleanCache, "cache", false, "Clean the Docker build cache")
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "Also clean all images unused after stopped containers are removed")
	cleanCmd.Flags().BoolVar(&cleanJSON, "json", false, "Output JSON")
	cleanCmd.Flags().DurationVar(&cleanTimeout, "timeout", 60*time.Second, "Timeout for Docker commands")
}

func validateCleanFlags(apply, all, images, containers, cache bool) error {
	hasIndividualScope := images || containers || cache
	if all && hasIndividualScope {
		return fmt.Errorf("clean: --all cannot be combined with --images, --containers, or --cache")
	}
	if apply && !all && !hasIndividualScope {
		return fmt.Errorf("clean: --apply requires --images, --containers, --cache, or --all")
	}
	return nil
}

func cleanScopeArgs(all, images, containers, cache bool) string {
	if all {
		return "--all"
	}

	args := make([]string, 0, 3)
	if images {
		args = append(args, "--images")
	}
	if containers {
		args = append(args, "--containers")
	}
	if cache {
		args = append(args, "--cache")
	}
	if len(args) == 0 {
		args = append(args, "--images", "--containers", "--cache")
	}
	return strings.Join(args, " ")
}

func sumImageBytes(items []clean.ImageItem) uint64 {
	var total uint64
	for _, item := range items {
		total += item.SizeBytes
	}
	return total
}

func sumContainerBytes(items []clean.ContainerItem) uint64 {
	var total uint64
	for _, item := range items {
		total += item.SizeBytes
	}
	return total
}

func formatBytes(bytes uint64) string {
	return sizeutil.Format(bytes)
}
