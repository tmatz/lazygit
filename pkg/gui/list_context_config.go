package gui

import (
	"log"

	"github.com/jesseduffield/lazygit/pkg/commands/git_commands"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/style"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

func (gui *Gui) menuListContext() *context.MenuContext {
	return context.NewMenuContext(
		gui.Views.Menu,
		nil,
		nil,
		nil,
		gui.c,
		gui.getMenuOptions,
	)
}

func (gui *Gui) filesListContext() *context.WorkingTreeContext {
	return context.NewWorkingTreeContext(
		func() []*models.File { return gui.State.Model.Files },
		gui.Views.Files,
		func(startIdx int, length int) [][]string {
			lines := presentation.RenderFileTree(gui.State.Contexts.Files.FileTreeViewModel, gui.State.Modes.Diffing.Ref, gui.State.Model.Submodules)
			mappedLines := make([][]string, len(lines))
			for i, line := range lines {
				mappedLines[i] = []string{line}
			}

			return mappedLines
		},
		OnFocusWrapper(gui.onFocusFile),
		OnFocusWrapper(gui.withDiffModeCheck(gui.filesRenderToMain)),
		nil,
		gui.c,
	)
}

func (gui *Gui) branchesListContext() *context.BranchesContext {
	return context.NewBranchesContext(
		func() []*models.Branch { return gui.State.Model.Branches },
		gui.Views.Branches,
		func(startIdx int, length int) [][]string {
			return presentation.GetBranchListDisplayStrings(gui.State.Model.Branches, gui.State.ScreenMode != SCREEN_NORMAL, gui.State.Modes.Diffing.Ref)
		},
		nil,
		OnFocusWrapper(gui.withDiffModeCheck(gui.branchesRenderToMain)),
		nil,
		gui.c,
	)
}

func (gui *Gui) remotesListContext() *context.RemotesContext {
	return context.NewRemotesContext(
		func() []*models.Remote { return gui.State.Model.Remotes },
		gui.Views.Branches,
		func(startIdx int, length int) [][]string {
			return presentation.GetRemoteListDisplayStrings(gui.State.Model.Remotes, gui.State.Modes.Diffing.Ref)
		},
		nil,
		OnFocusWrapper(gui.withDiffModeCheck(gui.remotesRenderToMain)),
		nil,
		gui.c,
	)
}

func (gui *Gui) remoteBranchesListContext() types.IListContext {
	return (&ListContext{
		BaseContext: context.NewBaseContext(context.NewBaseContextOpts{
			ViewName:   "branches",
			WindowName: "branches",
			Key:        context.REMOTE_BRANCHES_CONTEXT_KEY,
			Kind:       types.SIDE_CONTEXT,
			Focusable:  true,
		}),
		GetItemsLength:  func() int { return len(gui.State.Model.RemoteBranches) },
		OnGetPanelState: func() types.IListPanelState { return gui.State.Panels.RemoteBranches },
		OnRenderToMain:  OnFocusWrapper(gui.withDiffModeCheck(gui.remoteBranchesRenderToMain)),
		Gui:             gui,
		GetDisplayStrings: func(startIdx int, length int) [][]string {
			return presentation.GetRemoteBranchListDisplayStrings(gui.State.Model.RemoteBranches, gui.State.Modes.Diffing.Ref)
		},
		OnGetSelectedItemId: func() string {
			item := gui.getSelectedRemoteBranch()
			if item == nil {
				return ""
			}
			return item.ID()
		},
	}).attachKeybindings()
}

func (gui *Gui) withDiffModeCheck(f func() error) func() error {
	return func() error {
		if gui.State.Modes.Diffing.Active() {
			return gui.renderDiff()
		}

		return f()
	}
}

func (gui *Gui) tagsListContext() *context.TagsContext {
	return context.NewTagsContext(
		func() []*models.Tag { return gui.State.Model.Tags },
		gui.Views.Branches,
		func(startIdx int, length int) [][]string {
			return presentation.GetTagListDisplayStrings(gui.State.Model.Tags, gui.State.Modes.Diffing.Ref)
		},
		nil,
		OnFocusWrapper(gui.withDiffModeCheck(gui.tagsRenderToMain)),
		nil,
		gui.c,
	)
}

func (gui *Gui) branchCommitsListContext() *context.LocalCommitsContext {
	return context.NewLocalCommitsContext(
		func() []*models.Commit { return gui.State.Model.Commits },
		gui.Views.Commits,
		func(startIdx int, length int) [][]string {
			selectedCommitSha := ""
			if gui.currentContext().GetKey() == context.BRANCH_COMMITS_CONTEXT_KEY {
				selectedCommit := gui.State.Contexts.BranchCommits.GetSelectedCommit()
				if selectedCommit != nil {
					selectedCommitSha = selectedCommit.Sha
				}
			}
			return presentation.GetCommitListDisplayStrings(
				gui.State.Model.Commits,
				gui.State.ScreenMode != SCREEN_NORMAL,
				gui.helpers.CherryPick.CherryPickedCommitShaMap(),
				gui.State.Modes.Diffing.Ref,
				gui.c.UserConfig.Git.ParseEmoji,
				selectedCommitSha,
				startIdx,
				length,
				gui.shouldShowGraph(),
				gui.State.Model.BisectInfo,
			)
		},
		OnFocusWrapper(gui.onCommitFocus),
		OnFocusWrapper(gui.withDiffModeCheck(gui.branchCommitsRenderToMain)),
		nil,
		gui.c,
	)
}

func (gui *Gui) subCommitsListContext() types.IListContext {
	parseEmoji := gui.c.UserConfig.Git.ParseEmoji
	return (&ListContext{
		BaseContext: context.NewBaseContext(context.NewBaseContextOpts{
			ViewName:   "branches",
			WindowName: "branches",
			Key:        context.SUB_COMMITS_CONTEXT_KEY,
			Kind:       types.SIDE_CONTEXT,
			Focusable:  true,
		}),
		GetItemsLength:  func() int { return len(gui.State.Model.SubCommits) },
		OnGetPanelState: func() types.IListPanelState { return gui.State.Panels.SubCommits },
		OnRenderToMain:  OnFocusWrapper(gui.withDiffModeCheck(gui.subCommitsRenderToMain)),
		Gui:             gui,
		GetDisplayStrings: func(startIdx int, length int) [][]string {
			selectedCommitSha := ""
			if gui.currentContext().GetKey() == context.SUB_COMMITS_CONTEXT_KEY {
				selectedCommit := gui.getSelectedSubCommit()
				if selectedCommit != nil {
					selectedCommitSha = selectedCommit.Sha
				}
			}
			return presentation.GetCommitListDisplayStrings(
				gui.State.Model.SubCommits,
				gui.State.ScreenMode != SCREEN_NORMAL,
				gui.helpers.CherryPick.CherryPickedCommitShaMap(),
				gui.State.Modes.Diffing.Ref,
				parseEmoji,
				selectedCommitSha,
				startIdx,
				length,
				gui.shouldShowGraph(),
				git_commands.NewNullBisectInfo(),
			)
		},
		OnGetSelectedItemId: func() string {
			item := gui.getSelectedSubCommit()
			if item == nil {
				return ""
			}
			return item.ID()
		},
		RenderSelection: true,
	}).attachKeybindings()
}

func (gui *Gui) shouldShowGraph() bool {
	if gui.State.Modes.Filtering.Active() {
		return false
	}

	value := gui.c.UserConfig.Git.Log.ShowGraph
	switch value {
	case "always":
		return true
	case "never":
		return false
	case "when-maximised":
		return gui.State.ScreenMode != SCREEN_NORMAL
	}

	log.Fatalf("Unknown value for git.log.showGraph: %s. Expected one of: 'always', 'never', 'when-maximised'", value)
	return false
}

func (gui *Gui) reflogCommitsListContext() types.IListContext {
	parseEmoji := gui.c.UserConfig.Git.ParseEmoji
	return (&ListContext{
		BaseContext: context.NewBaseContext(context.NewBaseContextOpts{
			ViewName:   "commits",
			WindowName: "commits",
			Key:        context.REFLOG_COMMITS_CONTEXT_KEY,
			Kind:       types.SIDE_CONTEXT,
			Focusable:  true,
		}),
		GetItemsLength:  func() int { return len(gui.State.Model.FilteredReflogCommits) },
		OnGetPanelState: func() types.IListPanelState { return gui.State.Panels.ReflogCommits },
		OnRenderToMain:  OnFocusWrapper(gui.withDiffModeCheck(gui.reflogCommitsRenderToMain)),
		Gui:             gui,
		GetDisplayStrings: func(startIdx int, length int) [][]string {
			return presentation.GetReflogCommitListDisplayStrings(
				gui.State.Model.FilteredReflogCommits,
				gui.State.ScreenMode != SCREEN_NORMAL,
				gui.helpers.CherryPick.CherryPickedCommitShaMap(),
				gui.State.Modes.Diffing.Ref,
				parseEmoji,
			)
		},
		OnGetSelectedItemId: func() string {
			item := gui.getSelectedReflogCommit()
			if item == nil {
				return ""
			}
			return item.ID()
		},
	}).attachKeybindings()
}

func (gui *Gui) stashListContext() types.IListContext {
	return (&ListContext{
		BaseContext: context.NewBaseContext(context.NewBaseContextOpts{
			ViewName:   "stash",
			WindowName: "stash",
			Key:        context.STASH_CONTEXT_KEY,
			Kind:       types.SIDE_CONTEXT,
			Focusable:  true,
		}),
		GetItemsLength:  func() int { return len(gui.State.Model.StashEntries) },
		OnGetPanelState: func() types.IListPanelState { return gui.State.Panels.Stash },
		OnRenderToMain:  OnFocusWrapper(gui.withDiffModeCheck(gui.stashRenderToMain)),
		Gui:             gui,
		GetDisplayStrings: func(startIdx int, length int) [][]string {
			return presentation.GetStashEntryListDisplayStrings(gui.State.Model.StashEntries, gui.State.Modes.Diffing.Ref)
		},
		OnGetSelectedItemId: func() string {
			item := gui.getSelectedStashEntry()
			if item == nil {
				return ""
			}
			return item.ID()
		},
	}).attachKeybindings()
}

func (gui *Gui) commitFilesListContext() *context.CommitFilesContext {
	return context.NewCommitFilesContext(
		func() []*models.CommitFile { return gui.State.Model.CommitFiles },
		gui.Views.CommitFiles,
		func(startIdx int, length int) [][]string {
			if gui.State.Contexts.CommitFiles.CommitFileTreeViewModel.GetItemsLength() == 0 {
				return [][]string{{style.FgRed.Sprint("(none)")}}
			}

			lines := presentation.RenderCommitFileTree(gui.State.Contexts.CommitFiles.CommitFileTreeViewModel, gui.State.Modes.Diffing.Ref, gui.git.Patch.PatchManager)
			mappedLines := make([][]string, len(lines))
			for i, line := range lines {
				mappedLines[i] = []string{line}
			}

			return mappedLines
		},
		OnFocusWrapper(gui.onCommitFileFocus),
		OnFocusWrapper(gui.withDiffModeCheck(gui.commitFilesRenderToMain)),
		nil,
		gui.c,
	)
}

func (gui *Gui) submodulesListContext() *context.SubmodulesContext {
	return context.NewSubmodulesContext(
		func() []*models.SubmoduleConfig { return gui.State.Model.Submodules },
		gui.Views.Files,
		func(startIdx int, length int) [][]string {
			return presentation.GetSubmoduleListDisplayStrings(gui.State.Model.Submodules)
		},
		nil,
		OnFocusWrapper(gui.withDiffModeCheck(gui.submodulesRenderToMain)),
		nil,
		gui.c,
	)
}

func (gui *Gui) suggestionsListContext() types.IListContext {
	return (&ListContext{
		BaseContext: context.NewBaseContext(context.NewBaseContextOpts{
			ViewName:   "suggestions",
			WindowName: "suggestions",
			Key:        context.SUGGESTIONS_CONTEXT_KEY,
			Kind:       types.PERSISTENT_POPUP,
			Focusable:  true,
		}),
		GetItemsLength:  func() int { return len(gui.State.Suggestions) },
		OnGetPanelState: func() types.IListPanelState { return gui.State.Panels.Suggestions },
		Gui:             gui,
		GetDisplayStrings: func(startIdx int, length int) [][]string {
			return presentation.GetSuggestionListDisplayStrings(gui.State.Suggestions)
		},
	}).attachKeybindings()
}

func (gui *Gui) getListContexts() []types.IListContext {
	return []types.IListContext{
		gui.State.Contexts.Menu,
		gui.State.Contexts.Files,
		gui.State.Contexts.Branches,
		gui.State.Contexts.Remotes,
		gui.State.Contexts.RemoteBranches,
		gui.State.Contexts.Tags,
		gui.State.Contexts.BranchCommits,
		gui.State.Contexts.ReflogCommits,
		gui.State.Contexts.SubCommits,
		gui.State.Contexts.Stash,
		gui.State.Contexts.CommitFiles,
		gui.State.Contexts.Submodules,
		gui.State.Contexts.Suggestions,
	}
}
