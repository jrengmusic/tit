#pragma once
#include <JuceHeader.h>
#include "MenuItems.h"
#include "state/TitAxis.h"
#include <functional>
#include <unordered_map>

// ============================================================================
// MenuBuilder — MenuGeneratorMap dispatch
// ============================================================================
//
// Mirrors Go ___legacy___/internal/app/menu.go GenerateMenu():
//   map[git.Operation]MenuGenerator dispatch with Fail Fast on unknown key.
//
// RFC §3.6 contract:
//   DispatchTable[operation] → std::function<juce::Array<MenuItemDef>(const VT&)>
//
// BLESSED D — same repo VT + same operation → bit-identical MenuItemDef array.
// BLESSED L — single .at() lookup; zero switch chains.
// BLESSED S — no side effects; no member mutation; builder is const.
// BLESSED E — caller commits VT; builder returns data only.

namespace menu
{

class MenuBuilder
{
public:
    using Generator = std::function<juce::Array<MenuItemDef> (const juce::ValueTree& repoSubtree)>;

    MenuBuilder();

    // build() reads ID::operation from repoSubtree, dispatches to the registered
    // generator, and returns the resulting item array.
    // Precondition: repoSubtree must be valid and contain ID::operation.
    // Throws std::out_of_range (from .at()) if operation string does not map
    // to a registered Operation — Fail Fast per BLESSED E.
    juce::Array<MenuItemDef> build (const juce::ValueTree& repoSubtree) const;

private:
    std::unordered_map<Operation, Generator> generators;
};

} // namespace menu
