package app

import (
	"fmt"
	"io"
	"text/tabwriter"

	"faultline/internal/engine"
	"faultline/internal/output"
	"faultline/internal/playbooks"
	"faultline/internal/renderer"
)

type catalogService struct{}

func (catalogService) List(category, playbookDir string, playbookPacks []string, w io.Writer) error {
	pbs, err := engine.New(engine.Options{
		PlaybookDir:      playbookDir,
		PlaybookPackDirs: playbookPacks,
	}).List()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookList(pbs, category, renderer.DetectOptions(w)))
	return err
}

func (catalogService) Explain(id, playbookDir string, playbookPacks []string, format output.Format, w io.Writer) error {
	pb, err := engine.New(engine.Options{
		PlaybookDir:      playbookDir,
		PlaybookPackDirs: playbookPacks,
	}).Explain(id)
	if err != nil {
		return err
	}
	if format == output.FormatMarkdown {
		_, err = fmt.Fprint(w, output.FormatPlaybookDetailsMarkdown(pb))
		return err
	}
	if format == output.FormatJSON {
		data, err := output.FormatPlaybookDetailsJSON(pb)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, data)
		return err
	}
	_, err = fmt.Fprint(w, output.FormatPlaybookDetails(pb, renderer.DetectOptions(w)))
	return err
}

func (catalogService) ListInstalledPacks(w io.Writer) error {
	packs, err := playbooks.ListInstalledPacks()
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		_, err := fmt.Fprintln(w, "No installed playbook packs.")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tPLAYBOOKS\tVERSION\tPINNED REF\tPATH"); err != nil {
		return err
	}
	for _, pack := range packs {
		version := pack.Version
		if version == "" {
			version = "-"
		}
		ref := pack.PinnedRef
		if ref == "" {
			ref = "-"
		}
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\n", pack.Name, pack.PlaybookCount, version, ref, pack.Root); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (catalogService) InstallPack(srcDir, name string, force bool, w io.Writer) error {
	pack, err := playbooks.InstallPack(srcDir, name, force)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Installed pack %s with %d playbooks at %s\n", pack.Name, pack.PlaybookCount, pack.Root)
	return err
}
